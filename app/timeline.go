package app

import (
	"fmt"
	"sync"
	"termfedi/component"
	"termfedi/config"
	"termfedi/layer"
	"time"

	"github.com/gdamore/tcell/v2"
)

const timelineRefreshInterval = 20 * time.Second

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

	config config.Configuration
	showcw bool

	autoRefreshStarted bool
}

func NewTimelineScreen(cfg config.Configuration, client layer.Client) *TimelineScreen {
	ts := new(TimelineScreen)

	limit := 40
	if cfg.Session.Type == config.Misskey {
		limit = 100
	}
	ts.Timelines = newPagedList(cfg.UI, limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 0, ts.showcw)
	})
	ts.FetchLayer = client
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

	if !ts.autoRefreshStarted {
		ts.autoRefreshStarted = true
		time.AfterFunc(timelineRefreshInterval, func() { ts.autoRefresh(screen, ctx) })
	}

	ts.drawNotes(screen, ctx)
}

func (ts *TimelineScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	ts.resize(screen)
}

func (ts *TimelineScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {

	ts.drawNotes(screen, ctx)

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlP:
			ctx.Exit(0)

		case tcell.KeyCtrlR:
			ts.refreshData()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyCtrlE:
			switch ts.currunt_tl {
			case CURRUNTTL_GLOBAL:
				ts.currunt_tl = CURRUNTTL_LOCAL
			case CURRUNTTL_HOME:
				ts.currunt_tl = CURRUNTTL_GLOBAL
			case CURRUNTTL_LOCAL:
				ts.currunt_tl = CURRUNTTL_HOME
			}
			ts.refreshData()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyCtrlQ:
			ts.timelinelock.Lock()
			ts.showcw = !ts.showcw
			ts.Timelines.layout()
			ts.timelinelock.Unlock()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyCtrlX:
			ts.timelinelock.Lock()
			insertActionTargets(ts.Timelines.page())
			ts.timelinelock.Unlock()
			ctx.TranslateTo("action")

		case tcell.KeyCtrlN:
			ctx.TranslateTo("noti")

		case tcell.KeyLeft:
			ts.timelinelock.Lock()
			ts.Timelines.prev()
			ts.timelinelock.Unlock()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyRight:
			ts.timelinelock.Lock()
			ts.Timelines.next()
			ts.timelinelock.Unlock()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		}
	}
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

	time.AfterFunc(timelineRefreshInterval, func() { ts.autoRefresh(screen, ctx) })
}

func (ts *TimelineScreen) drawNotes(screen tcell.Screen, ctx ApplicationContext) {
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	for _, note := range ts.Timelines.page() {
		height := ts.Timelines.height(note)
		component.DrawNoteComponent(0, y, note, screen, textStyle, height, ts.showcw)
		y += height
	}

	page, total := ts.Timelines.position()
	footer := fmt.Sprintf(" %s Page %d/%d | [Quit] C-p | [Refresh] C-r | [ShowCW] C-q | [Notifications] C-n | [Actions] C-x | [Timeline] C-e | [Prev] <- [Next] -> ", curruntTimeline(ts.currunt_tl), page, total)

	ctx.DrawFooterbar(footer)
}
