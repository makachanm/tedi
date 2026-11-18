package app

import (
	"tedi/component"
	"tedi/config"
	"tedi/layer"
	utils "tedi/utils"

	"github.com/gdamore/tcell/v2"
)

var ActionTimelineNoteData []layer.Note
var actionStartSelection int

func init() {
	ActionTimelineNoteData = make([]layer.Note, 0)
}

func insertActionTargets(notes []layer.Note, selected int) {
	ActionTimelineNoteData = append(ActionTimelineNoteData, notes...)
	actionStartSelection = selected
}

func clearActionTargets() {
	ActionTimelineNoteData = make([]layer.Note, 0)
}

func actOnNote(client layer.Client, note layer.Note, react bool) string {
	switch {
	case react && client.PostReaction(note.Id):
		return "Reacted to @" + note.Author_finger
	case !react && client.PostRenote(note.Id):
		return "Renoted @" + note.Author_finger
	case react:
		return "Reaction failed"
	default:
		return "Renote failed"
	}
}

func (as *ActionScreen) leave(ctx ApplicationContext) {
	clearActionTargets()
	as.selection = 0
	ctx.TranslateTo("main")
}

type ActionType int

const (
	RENOTE ActionType = iota
	REACT
)

type ActionScreen struct {
	FetchLayer layer.Client

	config      config.Configuration
	action_mode ActionType
	selection   int
}

func NewActionScreen(cfg config.Configuration, client layer.Client) *ActionScreen {
	as := new(ActionScreen)
	as.FetchLayer = client

	as.action_mode = REACT
	as.config = cfg
	as.selection = 0

	return as
}

func (as *ActionScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	as.selection = min(actionStartSelection, max(len(ActionTimelineNoteData)-1, 0))
	as.drawNotes(screen, ctx)
}

func (as *ActionScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
}

func (as *ActionScreen) perform(mode ActionType, ctx ApplicationContext) {
	if as.selection >= len(ActionTimelineNoteData) {
		return
	}
	target := ActionTimelineNoteData[as.selection].Id
	switch mode {
	case REACT:
		as.FetchLayer.PostReaction(target)
	case RENOTE:
		as.FetchLayer.PostRenote(target)
	}
	as.leave(ctx)
}

func (as *ActionScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)
			return

		case "esc", "a":
			as.leave(ctx)
			return

		case "up", "k", "left", "h":
			if as.selection > 0 {
				as.selection--
			}

		case "down", "j", "right", "l":
			if as.selection < len(ActionTimelineNoteData)-1 {
				as.selection++
			}

		case "tab":
			if as.action_mode == REACT {
				as.action_mode = RENOTE
			} else {
				as.action_mode = REACT
			}

		case "enter":
			as.perform(as.action_mode, ctx)
			return
		case "x":
			as.perform(REACT, ctx)
			return
		case "c":
			as.perform(RENOTE, ctx)
			return
		}
	}

	as.drawNotes(screen, ctx)
}

func (as *ActionScreen) drawNotes(screen tcell.Screen, ctx ApplicationContext) {
	screen.Clear()

	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	screenW, screenH := screen.Size()
	y := 0
	for i, note := range ActionTimelineNoteData {
		if y >= screenH-1 {
			break
		}
		height := as.config.UI.MaxItemHeight
		if as.config.UI.Layout == config.LayoutFit {
			height = component.NoteHeight(note, screenW, 3, true)
		}
		if i == as.selection {
			utils.WriteTo(screen, 0, y, " > ", textStyle)
		}
		component.DrawNoteComponent(3, y, note, screen, textStyle, height, true)
		y += height
	}

	var selectedmode string
	switch as.action_mode {
	case REACT:
		selectedmode = "React"
	case RENOTE:
		selectedmode = "Renote"

	}

	footer := footerText(as.config.UI,
		"Action Mode: "+selectedmode,
		"[Esc] Back | [x] React | [c] Renote | [Enter] Do | [Tab] Mode | [Up/Down] Select")

	ctx.DrawFooterbar(footer)
	screen.Show()
}
