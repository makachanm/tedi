package app

import (
	"fmt"
	"sync"
	"termfedi/component"
	"termfedi/config"
	"termfedi/layer"
	"termfedi/utils"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/makachanm/flogger-lib"
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
	flogger.Printf("Getting timeline: %s", curruntTimeline(tl))
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
	Timelines    *utils.ItemAutoDemandPagination[layer.Note]
	timelinelock sync.Mutex

	FetchLayer layer.Client
	currunt_tl CurruntTL

	config config.Configuration
	showcw bool

	autoRefreshStarted bool
}

func NewTimelineScreen(cfg config.Configuration, client layer.Client) *TimelineScreen {
	ts := new(TimelineScreen)
	flogger.Println("Initializing TimelineScreen")

	switch cfg.Session.Type {
	case config.Misskey:
		ts.Timelines = utils.NewItemAutoDemandPagination[layer.Note](100, 5)
	default:
		ts.Timelines = utils.NewItemAutoDemandPagination[layer.Note](40, 5)
	}
	ts.FetchLayer = client

	ts.currunt_tl = CURRUNTTL_HOME
	items := getTimeline(ts.FetchLayer, ts.currunt_tl)
	flogger.Printf("Fetched %d items from timeline", len(items))
	for _, item := range items {
		ts.Timelines.PutItem(item)
	}

	ts.config = cfg
	ts.showcw = false

	return ts
}

func (ts *TimelineScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("TimelineScreen: InitScene")
	_, h := screen.Size()
	ts.Timelines.SetMaxItemPerPage(int(h / ts.config.UI.MaxItemHeight))

	if !ts.autoRefreshStarted {
		ts.autoRefreshStarted = true
		time.AfterFunc(timelineRefreshInterval, func() { ts.autoRefresh(screen, ctx) })
	}

	ts.drawNotes(screen, ctx)
}

func (ts *TimelineScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("TimelineScreen: WindowChangedScene")
	_, h := screen.Size()
	ts.Timelines.SetMaxItemPerPage(int(h / ts.config.UI.MaxItemHeight))
}

func (ts *TimelineScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {

	ts.drawNotes(screen, ctx)

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlP:
			flogger.Println("TimelineScreen: Ctrl+P pressed, exiting")
			ctx.Exit(0)

		case tcell.KeyCtrlR:
			flogger.Println("TimelineScreen: Ctrl+R pressed, refreshing data")
			ts.refreshData()

		case tcell.KeyCtrlE:
			flogger.Println("TimelineScreen: Ctrl+E pressed, changing timeline")
			switch ts.currunt_tl {
			case CURRUNTTL_GLOBAL:
				ts.currunt_tl = CURRUNTTL_LOCAL
			case CURRUNTTL_HOME:
				ts.currunt_tl = CURRUNTTL_GLOBAL
			case CURRUNTTL_LOCAL:
				ts.currunt_tl = CURRUNTTL_HOME
			}
			flogger.Printf("TimelineScreen: new timeline is %s", curruntTimeline(ts.currunt_tl))
			ts.refreshData()

		case tcell.KeyCtrlQ:
			flogger.Println("TimelineScreen: Ctrl+Q pressed, toggling CW")
			ts.showcw = !ts.showcw
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyCtrlX:
			flogger.Println("TimelineScreen: Ctrl+X pressed, switching to action scene")
			ts.timelinelock.Lock()
			insertActionTargets(ts.Timelines.GetCurruntPage())
			ts.timelinelock.Unlock()
			ctx.TranslateTo("action")

		case tcell.KeyCtrlN:
			flogger.Println("TimelineScreen: Ctrl+N pressed, switching to notification scene")
			ctx.TranslateTo("noti")

		case tcell.KeyLeft:
			flogger.Println("TimelineScreen: Left arrow pressed, going to previous page")
			ts.Timelines.GoPrev()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		case tcell.KeyRight:
			flogger.Println("TimelineScreen: Right arrow pressed, going to next page")
			ts.Timelines.GoNext()
			screen.Clear()
			ts.drawNotes(screen, ctx)

		}
	}
}

func (ts *TimelineScreen) refreshData() {
	flogger.Println("TimelineScreen: refreshing data")
	ts.timelinelock.Lock()
	defer ts.timelinelock.Unlock()
	currunt_pos := ts.Timelines.GetCurruntPagePointer()

	ts.Timelines.Clear()

	items := getTimeline(ts.FetchLayer, ts.currunt_tl)
	flogger.Printf("TimelineScreen: fetched %d new items", len(items))
	for _, item := range items {
		ts.Timelines.PutItem(item)
	}

	for i := 0; i < currunt_pos; i++ {
		ts.Timelines.GoNext()
	}

}

func (ts *TimelineScreen) autoRefresh(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("TimelineScreen: auto-refreshing data")
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

	items := ts.Timelines.GetCurruntPage()
	for i, notes := range items {
		component.DrawNoteComponent(0, i*ts.config.UI.MaxItemHeight, notes, screen, textStyle, ts.config.UI.MaxItemHeight, ts.showcw)
	}

	footer := fmt.Sprintf(" %s Page %d/%d | [Quit] C-p | [Refresh] C-r | [ShowCW] C-q | [Notifications] C-n | [Actions] C-x | [Timeline] C-e | [Prev] <- [Next] -> ", curruntTimeline(ts.currunt_tl), ts.Timelines.GetCurruntPagePointer()+1, ts.Timelines.GetTotalPage()+1)

	ctx.DrawFooterbar(footer)
}
