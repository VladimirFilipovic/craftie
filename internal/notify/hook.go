package notify

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/vlad/craftie/internal/config"
)

func Exec(dir string, cfgPath string) (bool, error) {
	if dir == "" {
		return false, fmt.Errorf("empty dir")
	}
	dir, _ = filepath.Abs(dir)

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return false, err
	}

	if !cfg.Notifications.Enabled {
		return false, fmt.Errorf("notifications not enabled")
	}

	// Match PWD against project paths
	var project string
	for name, path := range cfg.Notifications.Projects {
		abs, _ := filepath.Abs(path)
		if abs == dir {
			project = name
			break
		}
	}
	if project == "" {
		return false, nil
	}

	n, err := OpenBusNotifySession(OpenSessionParams{SoundEnabled: false})
	if err != nil {
		// NOTE: don't want to throw dbus related erros to avoid spamming the user
		// if something is indeed wrong when connecting to session
		// they can live without this notification
		return false, nil
	}
	defer n.Close()

	id, err := n.Send("Craftie",
		fmt.Sprintf("Start tracking %s?", project),
		[]string{"start", "Start Session", "dismiss", "Dismiss"})
	if err != nil {
		return false, nil
	}

	action, _ := n.WaitForAction(id, 15*time.Second)
	if action == "start" {
		return true, nil
	}
	return false, nil
}

// SetupScript returns shell integration script for the given shell type
func SetupScript(shell string) (string, error) {
	switch shell {
	case "zsh":
		return `craftie_chpwd() {
  craftie hook "$PWD" 2>/dev/null && print -z "craftie start -p $(basename $PWD) -e 60m &"
}
chpwd_functions+=craftie_chpwd
`, nil
	case "bash":
		return `__craftie_lastdir=""
__craftie_hook() {
  [ "$PWD" = "$__craftie_lastdir" ] && return
  __craftie_lastdir="$PWD"
  craftie hook "$PWD" 2>/dev/null && read -e -i "craftie start -p $(basename $PWD) -e 60m &" -p "> " cmd && eval "$cmd"
}
PROMPT_COMMAND="__craftie_hook;$PROMPT_COMMAND"
`, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (use zsh or bash)", shell)
	}
}
