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
	googlesheets "google.golang.org/api/sheets/v4"
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
					&cli.StringFlag{
						Name:     "task",
						Aliases:  []string{"t"},
						Usage:    "Task description",
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
	task := cmd.String("task")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🚀 Starting session for project:", projectName)
	fmt.Println("Configuration loaded")
	fmt.Println("Notes:", notes)
	fmt.Println("CFG:", cfg)

	var sheetsClient *googlesheets.Service
	if cfg.GoogleSheets.Enabled {
		var err error
		sheetsClient, err = sheets.NewSheetsClient(ctx, cfg.GoogleSheets.CredentialsHelper)
		if err != nil {
			return fmt.Errorf("failed to create Google Sheets client: %w", err)
		}
		fmt.Println("Google Sheets client created")
	}

	session := session.Session{
		StartTime:   time.Now(),
		Notes:       notes,
		ProjectName: projectName,
		Task:        task,
	}

	// Set up end timer if provided
	timerChan, err := session.SetEndTimer(endTimeStr)
	if err != nil {
		return err
	}

	syncChan := time.Tick(config.SessionSyncTime)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Started session for project \"%s\" have fun \n", projectName)

	saveParams := saveSessionParams{
		cfg:     cfg,
		session: &session,
		sheetsParams: sheets.GoogleSheetsParams{
			Srv:     sheetsClient,
			Cfg:     cfg.GoogleSheets,
			Session: &session,
		},
	}
	state := &syncState{}

	// Initial save
	saveSession(ctx, saveParams, state)

loop:
	for {
		select {
		case <-sigChan:
			fmt.Println("Session interrupted")
			break loop
		case <-timerChan:
			fmt.Println("Session time reached!")
			break loop
		case <-syncChan:
			fmt.Printf("Syncing session (duration: %s)\n", time.Time{}.Add(session.CurrentDuration()).Format(time.TimeOnly))
			saveSession(ctx, saveParams, state)
		}
	}

	session.Stop()

	fmt.Println("Session lasted ", time.Time{}.Add(session.CurrentDuration()).Format(time.TimeOnly))
	if session.Task != "" {
		fmt.Println("Task:", session.Task)
	}

	// Final sync to save end time
	saveSession(ctx, saveParams, state)

	return nil
}

type syncState struct {
	sheets *sheets.SyncState
	csv    *sheets.CsvSyncState
}

type saveSessionParams struct {
	cfg          *config.Config
	session      *session.Session
	sheetsParams sheets.GoogleSheetsParams
}

func saveSession(ctx context.Context, p saveSessionParams, state *syncState) {
	if p.cfg.CSV.Enabled {
		saveCsv(p, state)
	}
	if p.cfg.GoogleSheets.Enabled {
		saveGoogleSheets(ctx, p, state)
	}
}

func saveCsv(p saveSessionParams, state *syncState) {
	if state.csv != nil {
		if err := sheets.SyncCsvRow(state.csv, p.session); err != nil {
			fmt.Printf("Warning: failed to sync to CSV: %v\n", err)
		}
		return
	}

	csvState, err := sheets.InitCsvRow(p.cfg.CSV.FilePath, p.session)
	if err != nil {
		fmt.Printf("Warning: failed to init CSV row: %v\n", err)
		return
	}
	state.csv = csvState
	fmt.Printf("Session row created in CSV: %s\n", p.cfg.CSV.FilePath)
}

func saveGoogleSheets(ctx context.Context, p saveSessionParams, state *syncState) {
	// sheets already initialized
	if state.sheets != nil {
		if err := sheets.SyncGoogleSheetsRow(ctx, p.sheetsParams, state.sheets); err != nil {
			fmt.Printf("Warning: failed to sync to Google Sheets: %v\n", err)
		}
		return
	}

	// first time; need to init
	sheetsState, err := sheets.InitRow(ctx, p.sheetsParams)
	if err != nil {
		fmt.Printf("Warning: failed to init Google Sheets row: %v\n", err)
		return
	}
	state.sheets = sheetsState
	fmt.Printf("Session row created in Google Sheets (row %d)\n", sheetsState.RowNumber)
}
