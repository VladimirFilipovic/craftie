package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/session"
)

func main() {
	os.Exit(run())
}

func run() int {
	app := &cli.Command{
		Name:                   "craftie",
		Usage:                  "A time tracking application for crafters",
		UseShortOptionHandling: true,
		// TODO: read from git tree
		Version:        "0.0.1-beta",
		DefaultCommand: "start",
		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "Starts a new time tracking session. Stopping the previous active one.",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Aliases:  []string{"p"},
						Usage:    "Project name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config yaml file",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "notes",
						Aliases:  []string{"n"},
						Usage:    "Session notes",
						Required: false,
					},
				},
				Action: startSession,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

func startSession(ctx context.Context, cmd *cli.Command) error {
	projectName := cmd.String("project")
	notes := cmd.String("notes")
	configPath := cmd.String("config")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🚀 Starting session for project:", projectName)
	fmt.Println("Configuration loaded")
	fmt.Println("Notes:", notes)
	fmt.Println("CFG:", cfg)

	session := session.Session{
		StartTime:   time.Now(),
		Notes:       notes,
		ProjectName: projectName,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Started session for project \"%s\" have fun \n", projectName)

	<-sigChan

	fmt.Println("Session interrupted")
	fmt.Println("Session lasted ", session.FormattedDuration())

	return nil
}
