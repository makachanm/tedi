package app

import (
	"tedi/component"

	"github.com/gdamore/tcell/v2"
)

type menuEntry struct {
	label  string
	value  func() string
	run    func(ctx ApplicationContext)
	change func()
}

type MenuScreen struct {
	timeline *TimelineScreen
	profile  *ProfileScreen
	compose  *ComposeScreen
	cursor   int
	entries  []menuEntry
}

func NewMenuScreen(timeline *TimelineScreen, profile *ProfileScreen, compose *ComposeScreen) *MenuScreen {
	ms := &MenuScreen{timeline: timeline, profile: profile, compose: compose}

	timelineEntry := func(label string, tl CurruntTL) menuEntry {
		return menuEntry{
			label: label,
			value: func() string {
				if ms.timeline.currunt_tl == tl {
					return "showing"
				}
				return ""
			},
			run: func(ctx ApplicationContext) {
				ms.timeline.switchTimeline(tl)
				ctx.Back()
			},
		}
	}

	ms.entries = []menuEntry{
		timelineEntry("Home timeline", CURRUNTTL_HOME),
		timelineEntry("Local timeline", CURRUNTTL_LOCAL),
		timelineEntry("Global timeline", CURRUNTTL_GLOBAL),
		{
			label: "Write a note",
			run: func(ctx ApplicationContext) {
				ms.compose.open(nil)
				ctx.Replace("compose")
			},
		},
		{
			label: "Notifications",
			run:   func(ctx ApplicationContext) { ctx.Replace("noti") },
		},
		{
			label: "My profile",
			run: func(ctx ApplicationContext) {
				ms.profile.open("")
				ctx.Replace("profile")
			},
		},
		{
			label: "Refresh",
			run: func(ctx ApplicationContext) {
				ms.timeline.refreshData()
				ctx.Back()
			},
		},
		{
			label: "CW content",
			value: func() string {
				if ms.timeline.showcw {
					return "shown"
				}
				return "hidden"
			},
			change: func() { ms.timeline.toggleCW() },
		},
		{
			label: "Key guide",
			run:   func(ctx ApplicationContext) { ctx.Replace("main-help") },
		},
		{
			label: "Quit",
			run:   func(ctx ApplicationContext) { ctx.Exit(0) },
		},
	}
	return ms
}

func (ms *MenuScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	ms.cursor = 0
	screen.Clear()
	ms.draw(screen, ctx)
}

func (ms *MenuScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
}

func (ms *MenuScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		entry := ms.entries[ms.cursor]
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.Back()
			return

		case "f1":
			ctx.Replace("main-help")
			return

		case "up", "k":
			ms.cursor = max(ms.cursor-1, 0)
		case "down", "j":
			ms.cursor = min(ms.cursor+1, len(ms.entries)-1)

		case "left", "right", "h", "l":
			if entry.change != nil {
				entry.change()
			}

		case "enter":
			if entry.change != nil {
				entry.change()
				break
			}
			entry.run(ctx)
			return
		}
	}

	screen.Clear()
	ms.draw(screen, ctx)
}

func (ms *MenuScreen) draw(screen tcell.Screen, ctx ApplicationContext) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	items := make([]component.MenuItem, len(ms.entries))
	for i, e := range ms.entries {
		items[i].Label = e.label
		if e.value != nil {
			items[i].Value = e.value()
		}
		items[i].Choice = e.change != nil
	}
	component.DrawMenu(1, 1, items, ms.cursor, screen, style)

	ctx.DrawFooterbar(footerText("Menu", "Back"))
}
