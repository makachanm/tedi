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

type NotificationScreen struct {
	Notifications *pagedList[layer.Notification]
	notilock      sync.Mutex

	FetchLayer layer.Client
	detail     *NoteScreen

	config config.Configuration

	autoRefreshStarted bool
}

func NewNotificationScreen(cfg config.Configuration, client layer.Client, detail *NoteScreen) *NotificationScreen {
	ts := new(NotificationScreen)
	ts.Notifications = newPagedList(cfg.UI, 40, func(noti layer.Notification, screenW int) int {
		return component.NotificationHeight(noti, screenW, 1)
	})
	ts.FetchLayer = client
	ts.detail = detail
	ts.config = cfg

	ts.Notifications.setItems(ts.FetchLayer.GetNotifications())

	return ts
}

func (ns *NotificationScreen) resize(screen tcell.Screen) {
	ns.notilock.Lock()
	defer ns.notilock.Unlock()
	ns.Notifications.resize(screen.Size())
}

func (ns *NotificationScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	ns.resize(screen)

	if !ns.autoRefreshStarted && ns.config.UI.Refresh != config.RefreshOff {
		ns.autoRefreshStarted = true
		time.AfterFunc(refreshInterval(ns.config.UI), func() { ns.autoRefresh(screen, ctx) })
	}

	ns.drawNotis(screen, ctx)
}

func (ns *NotificationScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	ns.resize(screen)
}

func (ns *NotificationScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "m":
			ns.refreshData()

		case "esc":
			ctx.TranslateTo("main")
			return

		case "enter":
			ns.notilock.Lock()
			noti, ok := ns.Notifications.selected()
			ns.notilock.Unlock()
			if ok && noti.NoteId != "" {
				ns.detail.open(layer.Note{Id: noti.NoteId, Content: noti.Content}, "noti")
				ctx.TranslateTo("note")
				return
			}

		case "up", "k":
			ns.notilock.Lock()
			ns.Notifications.moveCursor(-1)
			ns.notilock.Unlock()

		case "down", "j":
			ns.notilock.Lock()
			ns.Notifications.moveCursor(1)
			ns.notilock.Unlock()

		case "left", "h", "pgup":
			ns.notilock.Lock()
			ns.Notifications.GoPrev()
			ns.notilock.Unlock()

		case "right", "l", "pgdn", " ":
			ns.notilock.Lock()
			ns.Notifications.GoNext()
			ns.notilock.Unlock()
		}
	}

	screen.Clear()
	ns.drawNotis(screen, ctx)
}

func (ns *NotificationScreen) refreshData() {
	items := ns.FetchLayer.GetNotifications()

	ns.notilock.Lock()
	defer ns.notilock.Unlock()
	ns.Notifications.setItems(items)
}

func (ns *NotificationScreen) autoRefresh(screen tcell.Screen, ctx ApplicationContext) {
	ns.refreshData()

	if ctx.GetCurruntScene() == "noti" {
		screen.Clear()
		ns.drawNotis(screen, ctx)
		screen.Show()
	}

	time.AfterFunc(refreshInterval(ns.config.UI), func() { ns.autoRefresh(screen, ctx) })
}

func (ns *NotificationScreen) drawNotis(screen tcell.Screen, ctx ApplicationContext) {
	ns.notilock.Lock()
	defer ns.notilock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	for i, noti := range ns.Notifications.GetCurruntPage() {
		height := ns.Notifications.height(noti)
		component.DrawNotificationComponent(1, y, noti, screen, textStyle, height, i == ns.Notifications.cursor)
		y += height
	}

	page, total := ns.Notifications.GetCurruntPagePointer()+1, ns.Notifications.GetTotalPage()+1
	footer := footerText(ns.config.UI,
		fmt.Sprintf("Noti Page %d/%d", page, total),
		"[Esc] Back | [Enter] Open | [m] Refresh | [Up/Down] Select | [<-/->] Page")

	ctx.DrawFooterbar(footer)
}
