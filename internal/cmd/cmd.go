package cmd

import (
	"context"

	"github.com/clementd64/codex/internal/app"
	"github.com/clementd64/codex/internal/igdb"
	"github.com/clementd64/codex/internal/tvshow"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Run the server",
				Action: func(ctx context.Context, c *cli.Command) error {
					app, err := app.New()
					if err != nil {
						return err
					}
					defer app.DB.Close()
					return server(ctx, app)
				},
			},
			{
				Name:  "cron",
				Usage: "Run background tasks",
				Action: func(ctx context.Context, c *cli.Command) error {
					app, err := app.New()
					if err != nil {
						return err
					}
					defer app.DB.Close()

					if err := tvshow.Fetch(ctx, app, tvshow.FetchOptions{
						Limit: 10,
					}); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "fetch",
				Usage: "fetch data",
				Commands: []*cli.Command{
					{
						Name: "tvshow",
						Flags: []cli.Flag{
							&cli.Int32Flag{
								Name:  "limit",
								Value: 10,
								Usage: "Number of show to refresh",
							},
							&cli.BoolFlag{
								Name:  "force",
								Usage: "Force refresh regardless of last sync time",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							app, err := app.New()
							if err != nil {
								return err
							}
							defer app.DB.Close()
							return tvshow.Fetch(ctx, app, tvshow.FetchOptions{
								Limit: c.Int32("limit"),
								Force: c.Bool("force"),
							})
						},
					},
					{
						Name: "games",
						Action: func(ctx context.Context, c *cli.Command) error {
							app, err := app.New()
							if err != nil {
								return err
							}
							defer app.DB.Close()
							return igdb.Fetch(ctx, app)
						},
					},
				},
			},
		},
	}

	return cmd.Run(ctx, args)
}
