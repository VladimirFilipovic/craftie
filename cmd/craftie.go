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
	"github.com/vlad/craftie/internal/sheets"
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
					&cli.StringFlag{
						Name:     "endtime",
						Aliases:  []string{"e"},
						Usage:    "Session end time duration (e.g., 2h, 30m, 1h30m)",
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
	// take flag values
	projectName := cmd.String("project")
	notes := cmd.String("notes")
	configPath := cmd.String("config")
	endTimeStr := cmd.String("endtime")

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

	// Set up end timer if provided
	timerChan, err := session.SetEndTimer(endTimeStr)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Started session for project \"%s\" have fun \n", projectName)

	select {
	case <-sigChan:
		fmt.Println("Session interrupted")
	case <-timerChan:
		fmt.Println("Session time reached!")
	}

	session.Stop()

	duration, err := session.Duration()

	if err != nil {
		return err
	}

	fmt.Println("Session lasted ", time.Time{}.Add(duration).Format(time.TimeOnly))

	// save session data
	if cfg.CSV.Enabled {
		if err := sheets.SaveToCsv(cfg.CSV.FilePath, &session); err != nil {
			fmt.Printf("Warning: failed to save session to CSV: %v\n", err)
		} else {
			fmt.Printf("Session saved to CSV: %s\n", cfg.CSV.FilePath)
		}
	}

	if cfg.GoogleSheets.Enabled {
		// save to google sheets
	}

	return nil
}
