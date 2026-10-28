package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/igdb"
)

func Run(ctx context.Context, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	app, err := app.New()
	if err != nil {
		return err
	}
	defer app.DB.Close()

	switch os.Args[1] {
	case "server":
		return server(ctx, app)
	case "fetch-igdb":
		return igdb.Fetch(ctx, app)
	default:
		usage()
		os.Exit(1)
	}

	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  codex server
  codex fetch-igdb`)
}
