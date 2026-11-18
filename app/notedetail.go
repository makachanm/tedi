package app

import (
	"fmt"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type NoteScreen struct {
	FetchLayer layer.Client
	config     config.Configuration

	note   layer.Note
	back   string
	notice string

	scroll int
	total  int
	rows   int
}

func NewNoteScreen(cfg config.Configuration, client layer.Client) *NoteScreen {
	ns := new(NoteScreen)
	ns.FetchLayer = client
	ns.config = cfg
	return ns
}

func (ns *NoteScreen) open(note layer.Note, back string) {
	ns.note = note
	ns.back = back
	ns.notice = ""
	ns.scroll = 0

	full, err := ns.FetchLayer.GetNote(note.Id)
	if err != nil {
		ns.notice = "Could not load full note"
		return
	}
	if full.RenotedBy == "" {
		full.RenotedBy = note.RenotedBy
	}
	ns.note = full
}

func (ns *NoteScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	screen.Clear()
	ns.drawNote(screen, ctx)
}

func (ns *NoteScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
}

func (ns *NoteScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		ns.notice = ""
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.TranslateTo(ns.back)
			return

		case "up", "k":
			ns.scroll--
		case "down", "j":
			ns.scroll++
		case "pgup", "left", "h":
			ns.scroll -= max(ns.rows-1, 1)
		case "pgdn", "right", "l", " ":
			ns.scroll += max(ns.rows-1, 1)

		case "x", "c":
			ns.notice = actOnNote(ns.FetchLayer, ns.note, keyName(ev) == "x")
		}
	}

	screen.Clear()
	ns.drawNote(screen, ctx)
}

func (ns *NoteScreen) drawNote(screen tcell.Screen, ctx ApplicationContext) {
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	w, h := screen.Size()

	ns.rows = h - 2
	ns.scroll, ns.total = component.DrawNoteDetail(2, 1, w-4, ns.rows, ns.note, screen, textStyle, ns.scroll)

	info := "Note"
	if ns.total > ns.rows {
		info += fmt.Sprintf(" %d-%d/%d", ns.scroll+1, min(ns.scroll+ns.rows, ns.total), ns.total)
	}
	if ns.notice != "" {
		info += " · " + ns.notice
	}
	footer := footerText(ns.config.UI, info,
		"[Esc] Back | [x] React | [c] Renote | [Up/Down] Scroll | [<-/->] Page")

	ctx.DrawFooterbar(footer)
}
