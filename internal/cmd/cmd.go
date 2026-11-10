package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/igdb"
	"github.com/clementd64/codex/internal/tvshow"
)

func Run(ctx context.Context, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(args) < 2 {
		usage()
		return fmt.Errorf("command required")
	}

	app, err := app.New()
	if err != nil {
		return err
	}
	defer app.DB.Close()

	switch args[1] {
	case "server":
		return server(ctx, app)
	case "fetch-igdb":
		return igdb.Fetch(ctx, app)
	case "fetch-tvshow":
		return tvshow.Fetch(ctx, app)
	default:
		usage()
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  codex server
  codex fetch-igdb
  codex fetch-tvshow`)
}
