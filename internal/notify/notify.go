package notify

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	dbusNotifDest = "org.freedesktop.Notifications"
	dbusNotifPath = "/org/freedesktop/Notifications"
	dbusNotifCall = "org.freedesktop.Notifications.Notify"
	defaultSound  = "/usr/share/sounds/freedesktop/stereo/complete.oga"

	dbusAddMatchCall = "org.freedesktop.DBus.AddMatch"
)

type Notifier struct {
	conn         *dbus.Conn
	soundEnabled bool
}

type OpenSessionParams struct {
	SoundEnabled bool
}

func OpenBusNotifySession(params OpenSessionParams) (*Notifier, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("dbus connect: %w", err)
	}
	return &Notifier{conn: conn, soundEnabled: params.SoundEnabled}, nil
}

func (n *Notifier) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

func (n *Notifier) Send(summary, body string, actions []string) (uint32, error) {
	if actions == nil {
		actions = []string{}
	}

	hints := map[string]dbus.Variant{}
	obj := n.conn.Object(dbusNotifDest, dbusNotifPath)
	call := obj.Call(dbusNotifCall, 0,
		"craftie", // app_name
		uint32(0), // replaces_id
		"",        // app_icon
		summary,   // summary
		body,      // body
		actions,   // actions
		hints,     // hints
		int32(-1), // expire_timeout (-1 = server default)
	)

	if call.Err != nil {
		return 0, call.Err
	}

	var id uint32
	if err := call.Store(&id); err != nil {
		return 0, err
	}

	if n.soundEnabled {
		exec.Command("paplay", defaultSound).Start()
	}

	return id, nil
}

// WaitForAction blocks until user clicks an action or timeout.
func (n *Notifier) WaitForAction(id uint32, timeout time.Duration) (string, error) {
	for _, member := range []string{"ActionInvoked", "NotificationClosed"} {
		rule := fmt.Sprintf("type='signal',interface='org.freedesktop.Notifications',member='%s'", member)
		if err := n.conn.BusObject().Call(dbusAddMatchCall, 0, rule).Err; err != nil {
			return "", err
		}
	}

	signals := make(chan *dbus.Signal, 2)
	n.conn.Signal(signals)
	defer n.conn.RemoveSignal(signals)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case sig := <-signals:
			if sig.Name == "org.freedesktop.Notifications.ActionInvoked" && len(sig.Body) >= 2 {
				sigID, ok1 := sig.Body[0].(uint32)
				action, ok2 := sig.Body[1].(string)
				if ok1 && ok2 && sigID == id {
					return action, nil
				}
			}
			// Notification closed without action
			if sig.Name == "org.freedesktop.Notifications.NotificationClosed" && len(sig.Body) >= 1 {
				if sigID, ok := sig.Body[0].(uint32); ok && sigID == id {
					fmt.Println("notification closed")
					return "", nil
				}
			}

			fmt.Println("unknown signal name:", sig)
			continue
		case <-timer.C:
			return "", fmt.Errorf("timed-out")
		}
	}
}
