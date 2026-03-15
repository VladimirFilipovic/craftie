package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/notify"
	"github.com/vlad/craftie/internal/session"
	"github.com/vlad/craftie/internal/sheets"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	app := &cli.Command{
		Name:                   "craftie",
		Usage:                  "A time tracking application for crafters",
		UseShortOptionHandling: true,
		Version:                version,
		DefaultCommand:         "start",
		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "Starts a new time tracking session. Stopping the previous active one.",
				Aliases: []string{"s"},
				Flags:   startFlags(),
				Action:  startCommand,
			},
			{
				Name:  "hook",
				Usage: "Shell hook: notify when entering a project directory",
				Commands: []*cli.Command{
					{
						Name:   "setup",
						Usage:  "Print shell integration snippet",
						Flags:  hookSetupFlags(),
						Action: hookSetupCommand,
					},
				},
				Flags:  hookFlags(),
				Action: hookCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func startFlags() []cli.Flag {
	return []cli.Flag{
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
		&cli.StringFlag{
			Name:     "task",
			Aliases:  []string{"t"},
			Usage:    "Task description",
			Required: false,
		},
	}
}

func startCommand(ctx context.Context, cmd *cli.Command) error {
	projectName := cmd.String("project")
	notes := cmd.String("notes")
	configPath := cmd.String("config")
	endTimeStr := cmd.String("endtime")
	task := cmd.String("task")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🚀 Starting session for project:", projectName)
	fmt.Println("Configuration loaded")

	sess := &session.Session{
		StartTime:   time.Now(),
		Notes:       notes,
		ProjectName: projectName,
		Task:        task,
	}

	sess.SetNotifications(cfg.Notifications.Enabled, cfg.Notifications.SoundEnabled)

	timerChan, err := sess.SetEndTimer(endTimeStr)
	if err != nil {
		return err
	}

	recorders := createRecorders(ctx, cfg)

	fmt.Printf("Started session for project \"%s\" have fun \n", projectName)

	if err := sess.Run(timerChan, config.SessionSyncTime, recorders); err != nil {
		return fmt.Errorf("session failed: %w", err)
	}

	fmt.Println("Session lasted ", sess.FormattedDuration())
	if sess.Task != "" {
		fmt.Println("Task:", sess.Task)
	}

	return nil
}

func createRecorders(ctx context.Context, cfg *config.Config) []session.Recorder {
	var recorders []session.Recorder

	if cfg.CSV.Enabled {
		recorders = append(recorders, sheets.NewCsvRecorder(cfg.CSV.FilePath))
	}

	if cfg.GoogleSheets.Enabled {
		if rec, err := sheets.NewGoogleSheetsRecorder(ctx, cfg.GoogleSheets); err == nil {
			recorders = append(recorders, rec)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: failed to create Google Sheets recorder: %v\n", err)
		}
	}

	return recorders
}

func hookFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Aliases:  []string{"c"},
			Usage:    "Path to config yaml file",
			Required: false,
		},
	}
}

func hookCommand(ctx context.Context, cmd *cli.Command) error {
	dir := cmd.Args().First()
	cfgPath := cmd.String("config")

	matched, err := notify.Exec(dir, cfgPath)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("no project match")
	}
	return nil
}

func hookSetupFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "shell",
			Aliases:  []string{"s"},
			Usage:    "Shell type (zsh or bash)",
			Required: true,
		},
	}
}

func hookSetupCommand(_ context.Context, cmd *cli.Command) error {
	shell := cmd.String("shell")
	script, err := notify.SetupScript(shell)
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}
