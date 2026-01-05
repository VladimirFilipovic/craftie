package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/session"
)

func main() {
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
		log.Fatal(err)
	}
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

	fmt.Println("Session in progress, have fun!")

	<-sigChan

	fmt.Println("Session ended manually.", session)
	fmt.Println("Duration:", session.GetDurationSec())

	return nil

	// Start session
	// startSession, err := sessionManager.StartSession(projectName, notes)
	// if err != nil {
	// 	return fmt.Errorf("failed to start session: %w", err)
	// }

	// fmt.Printf("✅ Started session for project '%s' (ID: %d)\n", projectName, startSession.ID)
	// fmt.Printf("   Started at: %s\n", startSession.StartTime.Format("2006-01-02 15:04:05"))
	// if notes != "" {
	// 	fmt.Printf("   Notes: %s\n", notes)
	// }

	// return nil
}

// func stopSession(ctx context.Context) error {
// 	// Initialize components
// 	configManager := config.NewManager()
// 	if err := configManager.Load(ctx.String("config")); err != nil {
// 		return fmt.Errorf("failed to load configuration: %w", err)
// 	}

// 	storage, err := storage.NewSQLiteStorage(configManager.GetConfig().Storage.DatabasePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to initialize storage: %w", err)
// 	}
// 	defer storage.Close()

// 	sessionManager := session.NewManager(storage)

// 	// Stop session
// 	stoppedSession, err := sessionManager.StopSession()
// 	if err != nil {
// 		return fmt.Errorf("failed to stop session: %w", err)
// 	}

// 	fmt.Printf("✅ Stopped session (ID: %d)\n", stoppedSession.ID)
// 	fmt.Printf("   Project: %s\n", stoppedSession.ProjectName)
// 	fmt.Printf("   Duration: %s\n", formatDuration(stoppedSession.Duration))
// 	fmt.Printf("   Started: %s\n", stoppedSession.StartTime.Format("2006-01-02 15:04:05"))
// 	fmt.Printf("   Ended: %s\n", stoppedSession.EndTime.Format("2006-01-02 15:04:05"))
// 	if stoppedSession.Notes != "" {
// 		fmt.Printf("   Notes: %s\n", stoppedSession.Notes)
// 	}

// 	return nil
// }

// func showStatus(ctx context.Context) error {
// 	// Initialize components
// 	configManager := config.NewManager()
// 	if err := configManager.Load(ctx.String("config")); err != nil {
// 		return fmt.Errorf("failed to load configuration: %w", err)
// 	}

// 	storage, err := storage.NewSQLiteStorage(configManager.GetConfig().Storage.DatabasePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to initialize storage: %w", err)
// 	}
// 	defer storage.Close()

// 	sessionManager := session.NewManager(storage)

// 	// Get status
// 	status, err := sessionManager.GetStatus()
// 	if err != nil {
// 		return fmt.Errorf("failed to get status: %w", err)
// 	}

// 	fmt.Println("📊 Craftie Status")
// 	fmt.Println("================")

// 	if status.IsActive {
// 		fmt.Printf("🟢 Active Session: %s\n", status.CurrentSession.ProjectName)
// 		fmt.Printf("   Started: %s\n", status.CurrentSession.StartTime.Format("2006-01-02 15:04:05"))
// func syncSheets(ctx context.Context) error {
// 	fmt.Println("🔄 Syncing with Google Sheets...")
// 	fmt.Println("This feature is not yet implemented.")
// 	return nil
// }
