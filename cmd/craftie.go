package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vlad/craftie/internal/config"
)

func main() {
	app := &cli.Command{
		Name:                   "craftie",
		Usage:                  "A time tracking application for crafters",
		UseShortOptionHandling: true,
		Version:                "0.0.1-beta",
		DefaultCommand:         "start",
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
					&cli.BoolFlag{
						Name:     "daemon",
						Aliases:  []string{"d"},
						Usage:    "Start craftie as a background service (recommended)",
						Required: false,
					},
				},
				Action: startSession,
			},
			// {
			// 	Name:    "stop",
			// 	Usage:   "Stop the current time tracking session",
			// 	Aliases: []string{"x"},
			// 	Action:  stopSession,
			// },
			// {
			// 	Name:    "status",
			// 	Usage:   "Show current tracking status",
			// 	Aliases: []string{"st"},
			// 	Action:  showStatus,
			// },
			// {
			// 	Name:    "list",
			// 	Usage:   "List recent sessions",
			// 	Aliases: []string{"ls"},
			// 	Flags: []cli.Flag{
			// 		&cli.IntFlag{
			// 			Name:    "limit",
			// 			Aliases: []string{"l"},
			// 			Usage:   "Number of sessions to show",
			// 			Value:   10,
			// 		},
			// 	},
			// 	Action: listSessions,
			// },
			// {
			// 	Name:    "sync",
			// 	Usage:   "Sync sessions with Google Sheets",
			// 	Aliases: []string{"sy"},
			// 	Action:  syncSheets,
			// },
			// {
			// 	Name:    "daemon",
			// 	Usage:   "Start the background daemon",
			// 	Aliases: []string{"d"},
			// 	Action:  startDaemon,
			// },
			// {
			// 	Name:    "exit",
			// 	Usage:   "Exit the daemon",
			// 	Aliases: []string{"e"},
			// 	Action:  exitDaemon,
			// },
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configuration file path",
			},
			&cli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   "Database file path",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Verbose output",
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

	configManager, err := config.NewConfigManager(
		configPath,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	fmt.Printf("🚀 Starting session for project '%s'...\n", projectName)
	fmt.Println("Configuration loaded from:", configPath)
	fmt.Printf("Print cfg %%v\n", configManager.Config)
	fmt.Println("Notes", notes)

	return nil

	// storage, err := storage.NewSQLiteStorage(configManager.Config.Storage.DatabasePath)
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize storage: %w", err)
	// }
	// defer storage.Close()

	// sessionManager := session.NewManager(storage)

	// // Start session
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
// 		fmt.Printf("   Duration: %s\n", formatDuration(status.CurrentSession.GetDuration()))
// 		if status.CurrentSession.Notes != "" {
// 			fmt.Printf("   Notes: %s\n", status.CurrentSession.Notes)
// 		}
// 	} else {
// 		fmt.Println("🔴 No active session")
// 	}

// 	fmt.Printf("\n📈 Statistics:\n")
// 	fmt.Printf("   Total Sessions: %d\n", status.TotalSessions)
// 	fmt.Printf("   Today's Sessions: %d\n", status.TodaySessions)
// 	fmt.Printf("   Today's Duration: %s\n", formatDuration(status.TodayDuration))

// 	return nil
// }

// func listSessions(ctx context.Context) error {
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

// 	// Get sessions
// 	filter := &types.SessionFilter{
// 		Limit: ctx.Int("limit"),
// 	}

// 	sessions, err := sessionManager.GetSessions(filter)
// 	if err != nil {
// 		return fmt.Errorf("failed to get sessions: %w", err)
// 	}

// 	fmt.Printf("📋 Recent Sessions (showing %d)\n", len(sessions))
// 	fmt.Println("================================")

// 	for _, session := range sessions {
// 		status := "✅"
// 		if session.IsActive() {
// 			status = "🟢"
// 		}

// 		fmt.Printf("%s %s - %s\n", status, session.StartTime.Format("2006-01-02 15:04"),
// 			formatDuration(session.GetDuration()))
// 		fmt.Printf("   Project: %s\n", session.ProjectName)
// 		if session.Notes != "" {
// 			fmt.Printf("   Notes: %s\n", session.Notes)
// 		}
// 		fmt.Println()
// 	}

// 	return nil
// }

// func syncSheets(ctx context.Context) error {
// 	fmt.Println("🔄 Syncing with Google Sheets...")
// 	fmt.Println("This feature is not yet implemented.")
// 	return nil
// }

// func startDaemon(ctx *cli.ct) error {
// 	fmt.Println("🚀 Starting daemon...")
// 	fmt.Println("This feature is not yet implemented.")
// 	return nil
// }

// func exitDaemon(ctx context.Context) error {
// 	fmt.Println("👋 Exiting daemon...")
// 	fmt.Println("This feature is not yet implemented.")
// 	return nil
// }

// func formatDuration(seconds uint64) string {
// 	if seconds < 60 {
// 		return fmt.Sprintf("%ds", seconds)
// 	}

// 	minutes := seconds / 60
// 	if minutes < 60 {
// 		return fmt.Sprintf("%dm %ds", minutes, seconds%60)
// 	}

// 	hours := minutes / 60
// 	minutes = minutes % 60
// 	return fmt.Sprintf("%dh %dm", hours, minutes)
// }
