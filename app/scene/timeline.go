package scene

import (
	"fmt"
	"sync"
	"tedi/app"
	"tedi/app/paging"
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
	Timelines    *paging.PagedList[layer.Note]
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

	ts.Timelines = paging.NewPagedList(cfg.UI, cfg.UI.Limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 1, ts.showcw)
	})
	ts.FetchLayer = client
	ts.detail = detail
	ts.profile = profile
	ts.compose = compose
	ts.config = cfg
	ts.showcw = false

	ts.currunt_tl = CURRUNTTL_HOME
	ts.Timelines.SetItems(getTimeline(ts.FetchLayer, ts.currunt_tl))

	return ts
}

func (ts *TimelineScreen) resize(screen tcell.Screen) {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	ts.Timelines.Resize(screen.Size())
}

func (ts *TimelineScreen) InitScene(screen tcell.Screen, ctx app.ApplicationContext) {
	ts.resize(screen)

	if !ts.autoRefreshStarted && ts.config.UI.Refresh != config.RefreshOff {
		ts.autoRefreshStarted = true
		time.AfterFunc(refreshInterval(ts.config.UI), func() { ts.autoRefresh(screen, ctx) })
	}

	ts.drawNotes(screen, ctx)
}

func (ts *TimelineScreen) WindowChangedScene(screen tcell.Screen, ctx app.ApplicationContext) {
	ts.resize(screen)
}

func (ts *TimelineScreen) toggleCW() {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	ts.showcw = !ts.showcw
	ts.Timelines.Layout()
}

func (ts *TimelineScreen) switchTimeline(tl CurruntTL) {
	ts.currunt_tl = tl
	ts.refreshData()
}

func (ts *TimelineScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx app.ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		ts.notice = ""
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.Forward("menu", nil)
			return

		case "up":
			ts.timelinelock.Lock()
			ts.Timelines.MoveCursor(-1)
			ts.timelinelock.Unlock()

		case "down":
			ts.timelinelock.Lock()
			ts.Timelines.MoveCursor(1)
			ts.timelinelock.Unlock()

		case "enter":
			ts.timelinelock.Lock()
			note, ok := ts.Timelines.Selected()
			ts.timelinelock.Unlock()
			if ok {
				ts.detail.open(note)
				ctx.Forward("note", nil)
				return
			}

		case "q":
			ts.refreshData()

		case "z":
			ts.toggleCW()

		case "tab":
			ctx.Forward("noti", nil)
			return

		case "left", "pgup":
			ts.timelinelock.Lock()
			ts.Timelines.GoPrev()
			ts.timelinelock.Unlock()

		case "right", "pgdn":
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
	ts.Timelines.SetItems(items)
}

func (ts *TimelineScreen) autoRefresh(screen tcell.Screen, ctx app.ApplicationContext) {
	ts.refreshData()

	if ctx.GetCurruntScene() == "main" {
		screen.Clear()
		ts.drawNotes(screen, ctx)
		screen.Show()
	}

	time.AfterFunc(refreshInterval(ts.config.UI), func() { ts.autoRefresh(screen, ctx) })
}

func (ts *TimelineScreen) drawNotes(screen tcell.Screen, ctx app.ApplicationContext) {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	for i, note := range ts.Timelines.GetCurruntPage() {
		height := ts.Timelines.Height(note)
		component.DrawNoteComponent(1, y, note, screen, textStyle, height, ts.showcw, i == ts.Timelines.Cursor)
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
