package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/auth"
	"github.com/clementd64/codex/internal/tvshow"
)

//go:generate go tool sqlc generate
//go:generate go tool buf generate

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app, err := app.New()
	if err != nil {
		return err
	}
	defer app.DB.Close()

	server := connect.NewServer(auth.Middleware(app.SigningKey))
	auth.Register(server, app)
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
	srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "Server failed", "error", err)
		os.Exit(1)
	}
}
