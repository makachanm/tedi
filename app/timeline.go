package app

import (
	"fmt"
	"sync"
	"tedi/component"
	"tedi/config"
	"tedi/layer"
	"time"

	"github.com/gdamore/tcell/v2"
)

type CurruntTL int

const (
	CURRUNTTL_GLOBAL CurruntTL = iota + 1
	CURRUNTTL_HOME
	CURRUNTTL_LOCAL
)

func curruntTimeline(tl CurruntTL) string {
	switch tl {
	case CURRUNTTL_GLOBAL:
		return "Glob"
	case CURRUNTTL_HOME:
		return "Home"
	case CURRUNTTL_LOCAL:
		return "Locl"
	default:
		return "Unknown"
	}
}

func getTimeline(layer layer.Client, tl CurruntTL) []layer.Note {
	switch tl {
	case CURRUNTTL_GLOBAL:
		return layer.GetGlobalTimeline()
	case CURRUNTTL_HOME:
		return layer.GetHomeTimeline()
	case CURRUNTTL_LOCAL:
		return layer.GetLocalTimeline()
	default:
		return nil
	}
}

type TimelineScreen struct {
	Timelines    *pagedList[layer.Note]
	timelinelock sync.Mutex

	FetchLayer layer.Client
	currunt_tl CurruntTL
	detail     *NoteScreen
	profile    *ProfileScreen
	compose    *ComposeScreen

	config config.Configuration
	showcw bool
	notice string

	autoRefreshStarted bool
}

func NewTimelineScreen(cfg config.Configuration, client layer.Client, detail *NoteScreen, profile *ProfileScreen, compose *ComposeScreen) *TimelineScreen {
	ts := new(TimelineScreen)

	ts.Timelines = newPagedList(cfg.UI, cfg.UI.Limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 1, ts.showcw)
	})
	ts.FetchLayer = client
	ts.detail = detail
	ts.profile = profile
	ts.compose = compose
	ts.config = cfg
	ts.showcw = false

	ts.currunt_tl = CURRUNTTL_HOME
	ts.Timelines.setItems(getTimeline(ts.FetchLayer, ts.currunt_tl))

	return ts
}

func (ts *TimelineScreen) resize(screen tcell.Screen) {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	ts.Timelines.resize(screen.Size())
}

func (ts *TimelineScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	ts.resize(screen)

	if !ts.autoRefreshStarted && ts.config.UI.Refresh != config.RefreshOff {
		ts.autoRefreshStarted = true
		time.AfterFunc(refreshInterval(ts.config.UI), func() { ts.autoRefresh(screen, ctx) })
	}

	ts.drawNotes(screen, ctx)
}

func (ts *TimelineScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	ts.resize(screen)
}

func (ts *TimelineScreen) toggleCW() {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	ts.showcw = !ts.showcw
	ts.Timelines.layout()
}

func (ts *TimelineScreen) switchTimeline(tl CurruntTL) {
	ts.currunt_tl = tl
	ts.refreshData()
}

func (ts *TimelineScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		ts.notice = ""
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.Forward("menu", nil)
			return

		case "q":
			ts.refreshData()

		case "f1":
			ctx.Forward("main-help", nil)
			return

		case "f2":
			ts.switchTimeline(CURRUNTTL_HOME)
		case "f3":
			ts.switchTimeline(CURRUNTTL_LOCAL)
		case "f4":
			ts.switchTimeline(CURRUNTTL_GLOBAL)

		case "z":
			ts.toggleCW()

		case "up", "k":
			ts.timelinelock.Lock()
			ts.Timelines.moveCursor(-1)
			ts.timelinelock.Unlock()

		case "down", "j":
			ts.timelinelock.Lock()
			ts.Timelines.moveCursor(1)
			ts.timelinelock.Unlock()

		case "x", "c":
			ts.timelinelock.Lock()
			note, ok := ts.Timelines.selected()
			ts.timelinelock.Unlock()
			if ok {
				ts.notice = actOnNote(ts.FetchLayer, note, keyName(ev) == "x")
			}

		case "enter":
			ts.timelinelock.Lock()
			note, ok := ts.Timelines.selected()
			ts.timelinelock.Unlock()
			if ok {
				ts.detail.open(note)
				ctx.Forward("note", nil)
				return
			}

		case "tab":
			ctx.Forward("noti", nil)
			return

		case "n":
			ts.compose.open(nil)
			ctx.Forward("compose", nil)
			return

		case "e":
			ts.profile.open("")
			ctx.Forward("profile", nil)
			return

		case "left", "h", "pgup":
			ts.timelinelock.Lock()
			ts.Timelines.GoPrev()
			ts.timelinelock.Unlock()

		case "right", "l", "pgdn", " ":
			ts.timelinelock.Lock()
			ts.Timelines.GoNext()
			ts.timelinelock.Unlock()
		}
	}

	screen.Clear()
	ts.drawNotes(screen, ctx)
}

func (ts *TimelineScreen) refreshData() {
	items := getTimeline(ts.FetchLayer, ts.currunt_tl)

	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	ts.Timelines.setItems(items)
}

func (ts *TimelineScreen) autoRefresh(screen tcell.Screen, ctx ApplicationContext) {
	ts.refreshData()

	if ctx.GetCurruntScene() == "main" {
		screen.Clear()
		ts.drawNotes(screen, ctx)
		screen.Show()
	}

	time.AfterFunc(refreshInterval(ts.config.UI), func() { ts.autoRefresh(screen, ctx) })
}

func (ts *TimelineScreen) drawNotes(screen tcell.Screen, ctx ApplicationContext) {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	for i, note := range ts.Timelines.GetCurruntPage() {
		height := ts.Timelines.height(note)
		component.DrawNoteComponent(1, y, note, screen, textStyle, height, ts.showcw, i == ts.Timelines.cursor)
		y += height
	}

	page, total := ts.Timelines.GetCurruntPagePointer()+1, ts.Timelines.GetTotalPage()+1
	info := fmt.Sprintf("%s Page %d/%d", curruntTimeline(ts.currunt_tl), page, total)
	if ts.notice != "" {
		info += " | " + ts.notice
	}
	ctx.DrawFooterbar(footerText(info, "Menu"))
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
