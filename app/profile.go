package app

import (
	"fmt"
	"sync"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type ProfileScreen struct {
	Notes    *pagedList[layer.Note]
	notelock sync.Mutex

	FetchLayer layer.Client
	detail     *NoteScreen
	config     config.Configuration

	profile layer.Profile
	header  int
	loaded  bool
	showcw  bool
	notice  string
	screenW int
	screenH int
}

func NewProfileScreen(cfg config.Configuration, client layer.Client, detail *NoteScreen) *ProfileScreen {
	ps := new(ProfileScreen)
	ps.Notes = newPagedList(cfg.UI, cfg.UI.Limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 1, ps.showcw)
	})
	ps.FetchLayer = client
	ps.detail = detail
	ps.config = cfg
	return ps
}

func (ps *ProfileScreen) resize(screen tcell.Screen) {
	ps.notelock.Lock()
	defer ps.notelock.Unlock()
	ps.screenW, ps.screenH = screen.Size()
	ps.header = component.ProfileHeight(ps.profile, ps.screenW, 2)
	ps.Notes.resize(ps.screenW, max(ps.screenH-ps.header, 2))
}

func (ps *ProfileScreen) refreshData(screen tcell.Screen) {
	profile, err := ps.FetchLayer.GetProfile()
	if err != nil {
		ps.notice = "Could not load profile"
		return
	}
	notes := ps.FetchLayer.GetUserNotes(profile.Id)

	ps.notelock.Lock()
	ps.profile = profile
	ps.loaded = true
	ps.Notes.setItems(notes)
	ps.notelock.Unlock()

	ps.resize(screen)
}

func (ps *ProfileScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	if !ps.loaded {
		ps.refreshData(screen)
	} else {
		ps.resize(screen)
	}

	screen.Clear()
	ps.drawProfile(screen, ctx)
}

func (ps *ProfileScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
	ps.resize(screen)
}

func (ps *ProfileScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		ps.notice = ""
		switch keyName(ev) {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			ctx.TranslateTo("main")
			return

		case "f1":
			ctx.TranslateTo("profile-help")
			return

		case "q":
			ps.refreshData(screen)

		case "z":
			ps.notelock.Lock()
			ps.showcw = !ps.showcw
			ps.Notes.layout()
			ps.notelock.Unlock()

		case "up", "k":
			ps.notelock.Lock()
			ps.Notes.moveCursor(-1)
			ps.notelock.Unlock()

		case "down", "j":
			ps.notelock.Lock()
			ps.Notes.moveCursor(1)
			ps.notelock.Unlock()

		case "x", "c":
			ps.notelock.Lock()
			note, ok := ps.Notes.selected()
			ps.notelock.Unlock()
			if ok {
				ps.notice = actOnNote(ps.FetchLayer, note, keyName(ev) == "x")
			}

		case "enter":
			ps.notelock.Lock()
			note, ok := ps.Notes.selected()
			ps.notelock.Unlock()
			if ok {
				ps.detail.open(note, "profile")
				ctx.TranslateTo("note")
				return
			}

		case "left", "h", "pgup":
			ps.notelock.Lock()
			ps.Notes.GoPrev()
			ps.notelock.Unlock()

		case "right", "l", "pgdn", " ":
			ps.notelock.Lock()
			ps.Notes.GoNext()
			ps.notelock.Unlock()
		}
	}

	screen.Clear()
	ps.drawProfile(screen, ctx)
}

func (ps *ProfileScreen) drawProfile(screen tcell.Screen, ctx ApplicationContext) {
	ps.notelock.Lock()
	defer ps.notelock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	if ps.loaded {
		y = component.DrawProfileHeader(2, 0, ps.profile, screen, textStyle)
	}

	for i, note := range ps.Notes.GetCurruntPage() {
		height := ps.Notes.height(note)
		component.DrawNoteComponent(1, y, note, screen, textStyle, height, ps.showcw, i == ps.Notes.cursor)
		y += height
	}

	page, total := ps.Notes.GetCurruntPagePointer()+1, ps.Notes.GetTotalPage()+1
	info := fmt.Sprintf("Profile Page %d/%d", page, total)
	if ps.notice != "" {
		info += " · " + ps.notice
	}
	ctx.DrawFooterbar(footerText(info))
}
