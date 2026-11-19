package scene

import (
	"tedi/app"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

type helpEntry struct {
	key  string
	desc string
}

type HelpScreen struct {
	title string
	keys  []helpEntry
}

func NewTimelineHelp() *HelpScreen {
	return &HelpScreen{title: "Timeline", keys: []helpEntry{
		{"Up / Down", "select note"},
		{"Left / Right", "previous / next page"},
		{"Enter", "open note"},
		{"Tab", "notifications"},
		{"Esc", "menu"},
	}}
}

func NewNotificationHelp() *HelpScreen {
	return &HelpScreen{title: "Notifications", keys: []helpEntry{
		{"Up / Down", "select notification"},
		{"Left / Right", "previous / next page"},
		{"Enter", "open note or user"},
		{"Tab / Esc", "back"},
	}}
}

func NewProfileHelp() *HelpScreen {
	return &HelpScreen{title: "Profile", keys: []helpEntry{
		{"Up / Down", "select note"},
		{"Left / Right", "previous / next page"},
		{"Enter", "open note"},
		{"Tab", "notifications"},
		{"Esc", "back"},
	}}
}

func NewNoteHelp() *HelpScreen {
	return &HelpScreen{title: "Note", keys: []helpEntry{
		{"Up / Down", "scroll"},
		{"PgUp / PgDn", "scroll a page"},
		{"Left / Right", "choose action"},
		{"Enter", "run action"},
		{"Tab", "notifications"},
		{"Esc", "back"},
	}}
}

func NewComposeHelp() *HelpScreen {
	return &HelpScreen{title: "Compose", keys: []helpEntry{
		{"Up / Down", "select item"},
		{"Left / Right", "change visibility"},
		{"Enter", "run item"},
		{"PgUp / PgDn", "scroll the preview"},
		{"Esc", "discard"},
	}}
}

func (hs *HelpScreen) InitScene(screen tcell.Screen, ctx app.ApplicationContext) {
	screen.Clear()
	hs.draw(screen, ctx)
}

func (hs *HelpScreen) WindowChangedScene(screen tcell.Screen, ctx app.ApplicationContext) {
}

func (hs *HelpScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx app.ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)
		case "esc", "f1":
			ctx.Back()
			return
		}
	}

	screen.Clear()
	hs.draw(screen, ctx)
}

func (hs *HelpScreen) draw(screen tcell.Screen, ctx app.ApplicationContext) {
	normal := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	dim := normal.Foreground(tcell.ColorGray)

	utils.WriteTo(screen, 2, 1, hs.title+" keys", normal.Bold(true))

	keyW := 0
	for _, e := range hs.keys {
		keyW = max(keyW, len(e.key))
	}
	for i, e := range hs.keys {
		utils.WriteTo(screen, 2, 3+i, e.key, normal)
		utils.WriteTo(screen, 2+keyW+3, 3+i, e.desc, dim)
	}

	ctx.DrawFooterbar(" " + hs.title + " help | [Esc] Back")
}
