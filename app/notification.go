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

const notificationRefreshInterval = 30 * time.Second

type NotificationScreen struct {
	Notifications *pagedList[layer.Notification]
	notilock      sync.Mutex

	FetchLayer layer.Client

	config config.Configuration

	autoRefreshStarted bool
}

func NewNotificationScreen(cfg config.Configuration, client layer.Client) *NotificationScreen {
	ts := new(NotificationScreen)
	ts.Notifications = newPagedList(cfg.UI, 40, func(noti layer.Notification, screenW int) int {
		return component.NotificationHeight(noti, screenW, 0)
	})
	ts.FetchLayer = client
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

	if !ns.autoRefreshStarted {
		ns.autoRefreshStarted = true
		time.AfterFunc(notificationRefreshInterval, func() { ns.autoRefresh(screen, ctx) })
	}

	ns.drawNotis(screen, ctx)
}

func (ns *NotificationScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	ns.resize(screen)
}

func (ns *NotificationScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	ns.drawNotis(screen, ctx)

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlP:
			ctx.Exit(0)

		case tcell.KeyCtrlR:
			ns.refreshData()
			screen.Clear()
			ns.drawNotis(screen, ctx)

		case tcell.KeyCtrlN:
			ctx.TranslateTo("main")

		case tcell.KeyLeft:
			ns.notilock.Lock()
			ns.Notifications.prev()
			ns.notilock.Unlock()
			screen.Clear()
			ns.drawNotis(screen, ctx)

		case tcell.KeyRight:
			ns.notilock.Lock()
			ns.Notifications.next()
			ns.notilock.Unlock()
			screen.Clear()
			ns.drawNotis(screen, ctx)

		}
	}
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

	time.AfterFunc(notificationRefreshInterval, func() { ns.autoRefresh(screen, ctx) })
}

func (ns *NotificationScreen) drawNotis(screen tcell.Screen, ctx ApplicationContext) {
	ns.notilock.Lock()
	defer ns.notilock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	for _, noti := range ns.Notifications.page() {
		height := ns.Notifications.height(noti)
		component.DrawNotificationComponent(0, y, noti, screen, textStyle, height)
		y += height
	}

	page, total := ns.Notifications.position()
	footer := fmt.Sprintf(" Noti Page %d/%d | [Quit] C-p | [Refresh] C-r | [Back] C-n | [Prev] <- [Next] -> ", page, total)

	ctx.DrawFooterbar(footer)
}
