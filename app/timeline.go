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

func (ts *TimelineScreen) switchTimeline(tl CurruntTL) {
	ts.currunt_tl = tl
	ts.refreshData()
}

func (ts *TimelineScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "q", "C-c":
			ctx.Exit(0)

		case "r":
			ts.refreshData()

		case "t":
			switch ts.currunt_tl {
			case CURRUNTTL_HOME:
				ts.switchTimeline(CURRUNTTL_LOCAL)
			case CURRUNTTL_LOCAL:
				ts.switchTimeline(CURRUNTTL_GLOBAL)
			default:
				ts.switchTimeline(CURRUNTTL_HOME)
			}
		case "1":
			ts.switchTimeline(CURRUNTTL_HOME)
		case "2":
			ts.switchTimeline(CURRUNTTL_LOCAL)
		case "3":
			ts.switchTimeline(CURRUNTTL_GLOBAL)

		case "c":
			ts.timelinelock.Lock()
			ts.showcw = !ts.showcw
			ts.Timelines.layout()
			ts.timelinelock.Unlock()

		case "a", "enter":
			ts.timelinelock.Lock()
			insertActionTargets(ts.Timelines.page())
			ts.timelinelock.Unlock()
			ctx.TranslateTo("action")

		case "n", "tab":
			ctx.TranslateTo("noti")

		case "left", "h", "pgup":
			ts.timelinelock.Lock()
			ts.Timelines.prev()
			ts.timelinelock.Unlock()

		case "right", "l", "pgdn", " ":
			ts.timelinelock.Lock()
			ts.Timelines.next()
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
	footer := footerText(ts.config.UI,
		fmt.Sprintf("%s Page %d/%d", curruntTimeline(ts.currunt_tl), page, total),
		"[q] Quit | [r] Refresh | [t/1-3] Timeline | [c] CW | [n] Notifications | [a] Actions | [<-/->] Page")

	ctx.DrawFooterbar(footer)
}
