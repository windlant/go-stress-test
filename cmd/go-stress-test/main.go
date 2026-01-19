package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/windlant/go-stress-test/internal/app"
	"github.com/windlant/go-stress-test/internal/config"
)

func main() {
	app := &cli.App{
		Name:  "go-stress-test",
		Usage: "A high-performance HTTP load testing tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Required: true,
				Usage:    "Target URL to stress test (e.g., http://localhost:8080/api)",
			},
			&cli.StringFlag{
				Name:    "method",
				Aliases: []string{"m"},
				Value:   "GET",
				Usage:   "HTTP method (GET, POST, PUT, DELETE, etc.)",
			},
			&cli.StringFlag{
				Name:  "body",
				Value: "",
				Usage: "Request body as a string",
			},
			&cli.StringFlag{
				Name:  "body-file",
				Value: "",
				Usage: "Path to file containing request body (overrides --body)",
			},
			&cli.StringSliceFlag{
				Name:    "header",
				Aliases: []string{"H"},
				Usage:   "Custom HTTP headers (can be used multiple times, e.g., -H \"Content-Type: application/json\")",
			},
			&cli.IntFlag{
				Name:    "rate",
				Aliases: []string{"r"},
				Value:   0,
				Usage:   "Requests per second (RPS)",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Value:   10,
				Usage:   "Number of concurrent workers (goroutines)",
			},
			&cli.DurationFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Value:   30 * 60 * time.Second,
				Usage:   "Duration of the test (e.g., 10s, 1m, 2h)",
			},
			&cli.IntFlag{
				Name:    "total",
				Aliases: []string{"n"},
				Value:   0,
				Usage:   "Total number of requests to send (overrides --duration if > 0)",
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Value: 30 * time.Second,
				Usage: "Timeout for each request",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "Output file for report (use 'stdout' for terminal)",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress progress output",
			},
			&cli.BoolFlag{
				Name:  "disable-keepalive",
				Usage: "Disable HTTP Keep-Alive connections (use short-lived connections)",
			},
		},
		Action: func(c *cli.Context) error {
			cfg := &config.Config{
				URL:              c.String("url"),
				Method:           c.String("method"),
				Body:             c.String("body"),
				BodyFile:         c.String("body-file"),
				Headers:          c.StringSlice("header"),
				Rate:             c.Int("rate"),
				Concurrency:      c.Int("concurrency"),
				Duration:         c.Duration("duration"),
				Total:            c.Int("total"),
				Timeout:          c.Duration("timeout"),
				Output:           c.String("output"),
				Quiet:            c.Bool("quiet"),
				DisableKeepalive: c.Bool("disable-keepalive"),
			}

			return app.Run(cfg)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(1)
	}
}
