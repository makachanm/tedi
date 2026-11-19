package scene

import (
	"fmt"
	"tedi/app"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type noteView struct {
	note      layer.Note
	ancestors []layer.Note
	replies   []layer.Note
	notice    string

	scroll int
	jump   bool
	action int
}

type noteAction int

const (
	actionReply noteAction = iota
	actionReact
	actionRenote
	actionAuthor
)

var noteActions = []string{"Reply", "React", "Renote", "Author"}

type NoteScreen struct {
	noteView

	FetchLayer layer.Client
	config     config.Configuration
	profile    *ProfileScreen
	compose    *ComposeScreen

	total int
	rows  int
}

func NewNoteScreen(cfg config.Configuration, client layer.Client) *NoteScreen {
	ns := new(NoteScreen)
	ns.FetchLayer = client
	ns.config = cfg
	return ns
}

func (ns *NoteScreen) SetScreens(profile *ProfileScreen, compose *ComposeScreen) {
	ns.profile = profile
	ns.compose = compose
}

func (ns *NoteScreen) snapshot() func() {
	saved := ns.noteView
	return func() { ns.noteView = saved }
}

func (ns *NoteScreen) open(note layer.Note) {
	ns.noteView = noteView{note: note, jump: true}

	full, err := ns.FetchLayer.GetNote(note.Id)
	if err != nil {
		ns.notice = "Could not load full note"
		return
	}
	if full.RenotedBy == "" {
		full.RenotedBy = note.RenotedBy
	}
	ns.note = full

	ns.ancestors, ns.replies, err = ns.FetchLayer.GetThread(note.Id)
	if err != nil {
		ns.notice = "Could not load the whole thread"
	}
}

func (ns *NoteScreen) InitScene(screen tcell.Screen, ctx app.ApplicationContext) {
	screen.Clear()
	ns.drawNote(screen, ctx)
}

func (ns *NoteScreen) WindowChangedScene(screen tcell.Screen, ctx app.ApplicationContext) {
}

func (ns *NoteScreen) runAction(action noteAction, ctx app.ApplicationContext) bool {
	switch action {
	case actionReply:
		if ns.note.Id == "" {
			return false
		}
		restore := ns.snapshot()
		reply := ns.note
		ns.compose.open(&reply)
		ctx.Forward("compose", restore)
		return true
	case actionReact:
		ns.notice = actOnNote(ns.FetchLayer, ns.note, true)
	case actionRenote:
		ns.notice = actOnNote(ns.FetchLayer, ns.note, false)
	case actionAuthor:
		if ns.note.Author_id == "" {
			return false
		}
		restore := ns.snapshot()
		ns.profile.open(ns.note.Author_id)
		ctx.Forward("profile", restore)
		return true
	}
	return false
}

func (ns *NoteScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx app.ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		ns.notice = ""
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.Back()
			return

		case "tab":
			ctx.Forward("noti", ns.snapshot())
			return

		case "left":
			ns.action = max(ns.action-1, 0)
		case "right":
			ns.action = min(ns.action+1, int(actionAuthor))

		case "enter":
			if ns.runAction(noteAction(ns.action), ctx) {
				return
			}

		case "up":
			ns.scroll--
		case "down":
			ns.scroll++
		case "pgup":
			ns.scroll -= max(ns.rows-1, 1)
		case "pgdn":
			ns.scroll += max(ns.rows-1, 1)
		}
	}

	screen.Clear()
	ns.drawNote(screen, ctx)
}

func (ns *NoteScreen) drawNote(screen tcell.Screen, ctx app.ApplicationContext) {
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	w, h := screen.Size()

	ns.rows = h - 4
	component.DrawActionBar(1, h-2, noteActions, ns.action, screen, textStyle)
	ns.scroll, ns.total = component.DrawNoteDetail(2, 1, w-4, ns.rows, ns.ancestors, ns.note, ns.replies, screen, textStyle, ns.scroll, ns.jump)
	ns.jump = false

	info := "Note"
	if n := len(ns.ancestors) + len(ns.replies); n > 0 {
		info += fmt.Sprintf(" (thread %d)", n+1)
	}
	if ns.total > ns.rows {
		info += fmt.Sprintf(" %d-%d/%d", ns.scroll+1, min(ns.scroll+ns.rows, ns.total), ns.total)
	}
	if ns.notice != "" {
		info += " | " + ns.notice
	}
	ctx.DrawFooterbar(footerText(info, "Back"))
}
