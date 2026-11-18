package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/auth"
	"github.com/clementd64/codex/internal/game"
	"github.com/clementd64/codex/internal/tvshow"
)

func server(ctx context.Context, application *app.App) error {
	server := connect.NewServer(auth.Middleware(application.SigningKey))
	auth.Register(server, application)
	game.Register(server, application)
	tvshowServer := tvshow.Register(server, application)

	api := http.NewServeMux()
	connecthttp.Mount(api, server)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.Handle("GET /ical/tvshow/{user}", app.IcalHandler(tvshowServer))
	mux.Handle("/", frontend())

	srv := &http.Server{
		Addr:     application.ListenAddr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	go func() {
		<-ctx.Done()
		slog.InfoContext(ctx, "Shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to shut down server", "error", err)
			_ = srv.Close()
		}
	}()

	slog.InfoContext(ctx, "Starting server", "addr", application.ListenAddr)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// frontendPath enables frontend serving when set at build time with
// -ldflags "-X github.com/clementd64/codex/internal/cmd.FrontPath=/path/to/dist".
var FrontPath string

func frontend() http.Handler {
	if FrontPath == "" {
		return http.NotFoundHandler()
	}

	files := noModTimeFS{os.DirFS(FrontPath)}
	fileServer := http.FileServerFS(files)

	// The immutable Nix store path identifies the whole frontend build.
	sum := sha256.Sum256([]byte(FrontPath))
	etag := `"` + base64.StdEncoding.EncodeToString(sum[:]) + `"`

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if _, err := fs.Stat(files, name); errors.Is(err, fs.ErrNotExist) {
			http.ServeFileFS(w, r, files, "index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

// noModTimeFS hides modification times: Nix normalizes store files to mtime 1.
// Publishing that Last-Modified date can make caches consider files fresh for
// years under heuristic freshness, without revalidating them.
type noModTimeFS struct{ fs.FS }

func (f noModTimeFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return noModTimeFile{file}, nil
}

type noModTimeFile struct{ fs.File }

func (f noModTimeFile) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return noModTimeInfo{info}, nil
}

type noModTimeInfo struct{ fs.FileInfo }

func (noModTimeInfo) ModTime() time.Time { return time.Time{} }
