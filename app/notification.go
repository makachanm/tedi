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

const notificationRefreshInterval = 30 * time.Second

type NotificationScreen struct {
	Notifications *utils.ItemAutoDemandPagination[layer.Notification]
	notilock      sync.Mutex

	FetchLayer layer.Client

	config config.Configuration

	autoRefreshStarted bool
}

func NewNotificationScreen(cfg config.Configuration, client layer.Client) *NotificationScreen {
	ts := new(NotificationScreen)
	flogger.Println("Initializing NotificationScreen")
	ts.Notifications = utils.NewItemAutoDemandPagination[layer.Notification](40, 5)
	ts.FetchLayer = client

	items := ts.FetchLayer.GetNotifications()
	flogger.Printf("Fetched %d notifications", len(items))
	for _, item := range items {
		ts.Notifications.PutItem(item)
	}

	ts.config = cfg

	return ts
}

func (ns *NotificationScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("NotificationScreen: InitScene")
	_, h := screen.Size()
	ns.Notifications.SetMaxItemPerPage(int(h / ns.config.UI.MaxItemHeight))

	if !ns.autoRefreshStarted {
		ns.autoRefreshStarted = true
		time.AfterFunc(notificationRefreshInterval, func() { ns.autoRefresh(screen, ctx) })
	}

	ns.drawNotis(screen, ctx)
}

func (ns *NotificationScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("NotificationScreen: WindowChangedScene")
	_, h := screen.Size()
	ns.Notifications.SetMaxItemPerPage(int(h / ns.config.UI.MaxItemHeight))
}

func (ns *NotificationScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	ns.drawNotis(screen, ctx)

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlP:
			flogger.Println("NotificationScreen: Ctrl+P pressed, exiting")
			ctx.Exit(0)

		case tcell.KeyCtrlR:
			flogger.Println("NotificationScreen: Ctrl+R pressed, refreshing data")
			ns.refreshData()

		case tcell.KeyCtrlN:
			flogger.Println("NotificationScreen: Ctrl+N pressed, switching to main scene")
			ctx.TranslateTo("main")

		case tcell.KeyLeft:
			flogger.Println("NotificationScreen: Left arrow pressed, going to previous page")
			ns.Notifications.GoPrev()
			screen.Clear()
			ns.drawNotis(screen, ctx)

		case tcell.KeyRight:
			flogger.Println("NotificationScreen: Right arrow pressed, going to next page")
			ns.Notifications.GoNext()
			screen.Clear()
			ns.drawNotis(screen, ctx)

		}
	}
}

func (ns *NotificationScreen) refreshData() {
	flogger.Println("NotificationScreen: refreshing data")
	ns.notilock.Lock()
	defer ns.notilock.Unlock()
	currunt_pos := ns.Notifications.GetCurruntPagePointer()

	ns.Notifications.Clear()

	items := ns.FetchLayer.GetNotifications()
	flogger.Printf("NotificationScreen: fetched %d new notifications", len(items))
	for _, item := range items {
		ns.Notifications.PutItem(item)
	}

	for i := 0; i < currunt_pos; i++ {
		ns.Notifications.GoNext()
	}

}

func (ns *NotificationScreen) autoRefresh(screen tcell.Screen, ctx ApplicationContext) {
	flogger.Println("NotificationScreen: auto-refreshing data")
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

	items := ns.Notifications.GetCurruntPage()
	for i, notes := range items {
		component.DrawNotificationComponent(0, i*ns.config.UI.MaxItemHeight, notes, screen, textStyle, ns.config.UI.MaxItemHeight)
	}

	footer := fmt.Sprintf(" Noti Page %d/%d | [Quit] C-p | [Refresh] C-r | [Back] C-n | [Prev] <- [Next] -> ", ns.Notifications.GetCurruntPagePointer()+1, ns.Notifications.GetTotalPage()+1)

	ctx.DrawFooterbar(footer)
}
