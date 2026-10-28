package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/clementd64/codex/internal/cmd"
)

//go:generate go tool sqlc generate
//go:generate go tool buf generate

func main() {
	ctx := context.Background()
	if err := cmd.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, "Server failed", "error", err)
		os.Exit(1)
	}
}
