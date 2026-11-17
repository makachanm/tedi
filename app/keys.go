package app

import "github.com/gdamore/tcell/v2"

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
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyEscape:
		return "esc"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyCtrlC:
		return "C-c"
	}
	return ""
}
