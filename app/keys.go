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
	case tcell.KeyEscape, tcell.KeyBackspace, tcell.KeyBackspace2:
		return "esc"
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyF1:
		return "f1"
	case tcell.KeyF2:
		return "f2"
	case tcell.KeyF3:
		return "f3"
	case tcell.KeyF4:
		return "f4"
	case tcell.KeyCtrlC:
		return "C-c"
	}
	return ""
}
