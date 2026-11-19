package scene

import (
	"tedi/app"
	"tedi/component"

	"github.com/gdamore/tcell/v2"
)

type menuItem int

const (
	menuHome menuItem = iota
	menuLocal
	menuGlobal
	menuWrite
	menuNotifications
	menuProfile
	menuQuit
)

type MenuScreen struct {
	timeline *TimelineScreen
	profile  *ProfileScreen
	compose  *ComposeScreen
	cursor   int
}

func NewMenuScreen(timeline *TimelineScreen, profile *ProfileScreen, compose *ComposeScreen) *MenuScreen {
	ms := new(MenuScreen)
	ms.timeline = timeline
	ms.profile = profile
	ms.compose = compose
	return ms
}

func (ms *MenuScreen) item(item menuItem) component.MenuItem {
	switch item {
	case menuHome:
		return component.MenuItem{Label: "Home timeline", Value: ms.showing(CURRUNTTL_HOME)}
	case menuLocal:
		return component.MenuItem{Label: "Local timeline", Value: ms.showing(CURRUNTTL_LOCAL)}
	case menuGlobal:
		return component.MenuItem{Label: "Global timeline", Value: ms.showing(CURRUNTTL_GLOBAL)}
	case menuWrite:
		return component.MenuItem{Label: "Write a note"}
	case menuNotifications:
		return component.MenuItem{Label: "Notifications"}
	case menuProfile:
		return component.MenuItem{Label: "My profile"}
	case menuQuit:
		return component.MenuItem{Label: "Quit"}
	}
	return component.MenuItem{}
}

func (ms *MenuScreen) showing(tl CurruntTL) string {
	if ms.timeline.currunt_tl == tl {
		return "showing"
	}
	return ""
}

func (ms *MenuScreen) run(item menuItem, ctx app.ApplicationContext) {
	switch item {
	case menuHome:
		ms.timeline.switchTimeline(CURRUNTTL_HOME)
		ctx.Back()
	case menuLocal:
		ms.timeline.switchTimeline(CURRUNTTL_LOCAL)
		ctx.Back()
	case menuGlobal:
		ms.timeline.switchTimeline(CURRUNTTL_GLOBAL)
		ctx.Back()
	case menuWrite:
		ms.compose.open(nil)
		ctx.Replace("compose")
	case menuNotifications:
		ctx.Replace("noti")
	case menuProfile:
		ms.profile.open("")
		ctx.Replace("profile")
	case menuQuit:
		ctx.Exit(0)
	}
}

func (ms *MenuScreen) InitScene(screen tcell.Screen, ctx app.ApplicationContext) {
	ms.cursor = 0
	screen.Clear()
	ms.draw(screen, ctx)
}

func (ms *MenuScreen) WindowChangedScene(screen tcell.Screen, ctx app.ApplicationContext) {
}

func (ms *MenuScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx app.ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.Back()
			return

		case "up":
			if ms.cursor > 0 {
				ms.cursor--
			}
		case "down":
			if ms.cursor < int(menuQuit) {
				ms.cursor++
			}

		case "enter":
			ms.run(menuItem(ms.cursor), ctx)
			return
		}
	}

	screen.Clear()
	ms.draw(screen, ctx)
}

func (ms *MenuScreen) draw(screen tcell.Screen, ctx app.ApplicationContext) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	var items []component.MenuItem
	for item := menuHome; item <= menuQuit; item++ {
		items = append(items, ms.item(item))
	}
	component.DrawMenu(1, 1, items, ms.cursor, screen, style)

	ctx.DrawFooterbar(footerText("Menu", "[Esc] Back"))
}
