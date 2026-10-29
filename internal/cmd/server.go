package cmd

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/auth"
	"github.com/clementd64/codex/internal/game"
	"github.com/clementd64/codex/internal/tvshow"
)

func server(ctx context.Context, app *app.App) error {
	server := connect.NewServer(auth.Middleware(app.SigningKey))
	auth.Register(server, app)
	game.Register(server, app)
	tvshow.Register(server, app)

	api := http.NewServeMux()
	connecthttp.Mount(api, server)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.Handle("/", frontend())

	srv := &http.Server{
		Addr:     app.ListenAddr,
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

	slog.InfoContext(ctx, "Starting server", "addr", app.ListenAddr)
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

	files := os.DirFS(FrontPath)

	fileServer := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(files, name); errors.Is(err, fs.ErrNotExist) {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
