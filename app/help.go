package app

import (
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

type helpEntry struct {
	key  string
	desc string
}

type HelpScreen struct {
	title string
	back  string
	keys  []helpEntry
}

func NewTimelineHelp() *HelpScreen {
	return &HelpScreen{title: "Timeline", back: "main", keys: []helpEntry{
		{"Up / Down, k / j", "select note"},
		{"Left / Right, h / l", "previous / next page"},
		{"Enter", "open note and its thread"},
		{"x", "react to selected note"},
		{"c", "renote selected note"},
		{"z", "show or hide CW content"},
		{"F2  F3  F4", "home / local / global timeline"},
		{"q", "refresh"},
		{"Tab", "notifications"},
		{"F1", "this guide"},
		{"Esc", "quit"},
	}}
}

func NewNotificationHelp() *HelpScreen {
	return &HelpScreen{title: "Notifications", back: "noti", keys: []helpEntry{
		{"Up / Down, k / j", "select notification"},
		{"Left / Right, h / l", "previous / next page"},
		{"Enter", "open the note it refers to"},
		{"q", "refresh"},
		{"F1", "this guide"},
		{"Esc", "back to timeline"},
	}}
}

func NewNoteHelp() *HelpScreen {
	return &HelpScreen{title: "Note", back: "note", keys: []helpEntry{
		{"Up / Down, k / j", "scroll"},
		{"Left / Right, h / l", "scroll a page"},
		{"x", "react to this note"},
		{"c", "renote this note"},
		{"F1", "this guide"},
		{"Esc", "back"},
	}}
}

func (hs *HelpScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	screen.Clear()
	hs.draw(screen, ctx)
}

func (hs *HelpScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
}

func (hs *HelpScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)
		case "esc", "f1":
			ctx.TranslateTo(hs.back)
			return
		}
	}

	screen.Clear()
	hs.draw(screen, ctx)
}

func (hs *HelpScreen) draw(screen tcell.Screen, ctx ApplicationContext) {
	normal := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	dim := normal.Foreground(tcell.ColorGray)

	utils.WriteTo(screen, 2, 1, hs.title+" keys", normal.Bold(true))

	keyW := 0
	for _, e := range hs.keys {
		keyW = max(keyW, len(e.key))
	}
	for i, e := range hs.keys {
		utils.WriteTo(screen, 4, 3+i, e.key, normal)
		utils.WriteTo(screen, 4+keyW+3, 3+i, e.desc, dim)
	}

	ctx.DrawFooterbar(" Help · " + hs.title + " | [Esc] Back")
}
