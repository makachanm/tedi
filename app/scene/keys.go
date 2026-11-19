package scene

import (
	"tedi/config"
	"time"

	"github.com/gdamore/tcell/v2"
)

func refreshInterval(ui config.UIConfiguration) time.Duration {
	return time.Duration(ui.Refresh) * time.Second
}

func footerText(info string, keys string) string {
	return " " + info + " | " + keys
}

func keyName(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyRune:
		return string(ev.Rune())
	case tcell.KeyLeft:
		return "left"
	case tcell.KeyRight:
		return "right"
	case tcell.KeyUp:
		return "up"
	case tcell.KeyDown:
		return "down"
	case tcell.KeyPgUp:
		return "pgup"
	case tcell.KeyPgDn:
		return "pgdn"
	case tcell.KeyEscape, tcell.KeyBackspace, tcell.KeyBackspace2:
		return "esc"
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyCtrlC:
		return "C-c"
	}
	return ""
}
