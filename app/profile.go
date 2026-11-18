package app

import (
	"fmt"
	"sync"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type profileView struct {
	userId  string
	profile layer.Profile
	notes   *pagedList[layer.Note]
	loaded  bool
	showcw  bool
	notice  string
}

type ProfileScreen struct {
	profileView
	lock sync.Mutex

	FetchLayer layer.Client
	detail     *NoteScreen
	config     config.Configuration
}

func NewProfileScreen(cfg config.Configuration, client layer.Client, detail *NoteScreen) *ProfileScreen {
	ps := new(ProfileScreen)
	ps.FetchLayer = client
	ps.detail = detail
	ps.config = cfg
	ps.open("")
	return ps
}

func (ps *ProfileScreen) open(userId string) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	ps.profileView = profileView{userId: userId}
	ps.notes = newPagedList(ps.config.UI, ps.config.UI.Limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 1, ps.showcw)
	})
}

func (ps *ProfileScreen) snapshot() func() {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	saved := ps.profileView
	return func() {
		ps.lock.Lock()
		defer ps.lock.Unlock()
		ps.profileView = saved
	}
}

func (ps *ProfileScreen) resize(screen tcell.Screen) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	w, h := screen.Size()
	header := component.ProfileHeight(ps.profile, w, 2)
	ps.notes.resize(w, max(h-header, 2))
}

func (ps *ProfileScreen) refreshData(screen tcell.Screen) {
	var profile layer.Profile
	var err error
	if ps.userId == "" {
		profile, err = ps.FetchLayer.GetProfile()
	} else {
		profile, err = ps.FetchLayer.GetUser(ps.userId)
	}
	if err != nil {
		ps.notice = "Could not load profile"
		return
	}
	notes := ps.FetchLayer.GetUserNotes(profile.Id)

	ps.lock.Lock()
	ps.profile = profile
	ps.loaded = true
	ps.notes.setItems(notes)
	ps.lock.Unlock()

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
			ctx.Back()
			return

		case "f1":
			ctx.Forward("profile-help", nil)
			return

		case "q":
			ps.refreshData(screen)

		case "z":
			ps.lock.Lock()
			ps.showcw = !ps.showcw
			ps.notes.layout()
			ps.lock.Unlock()

		case "up", "k":
			ps.lock.Lock()
			ps.notes.moveCursor(-1)
			ps.lock.Unlock()

		case "down", "j":
			ps.lock.Lock()
			ps.notes.moveCursor(1)
			ps.lock.Unlock()

		case "x", "c":
			ps.lock.Lock()
			note, ok := ps.notes.selected()
			ps.lock.Unlock()
			if ok {
				ps.notice = actOnNote(ps.FetchLayer, note, keyName(ev) == "x")
			}

		case "enter":
			ps.lock.Lock()
			note, ok := ps.notes.selected()
			ps.lock.Unlock()
			if ok {
				restore := ps.snapshot()
				ps.detail.open(note)
				ctx.Forward("note", restore)
				return
			}

		case "left", "h", "pgup":
			ps.lock.Lock()
			ps.notes.GoPrev()
			ps.lock.Unlock()

		case "right", "l", "pgdn", " ":
			ps.lock.Lock()
			ps.notes.GoNext()
			ps.lock.Unlock()
		}
	}

	screen.Clear()
	ps.drawProfile(screen, ctx)
}

func (ps *ProfileScreen) drawProfile(screen tcell.Screen, ctx ApplicationContext) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	if ps.loaded {
		y = component.DrawProfileHeader(2, 0, ps.profile, screen, textStyle)
	}

	for i, note := range ps.notes.GetCurruntPage() {
		height := ps.notes.height(note)
		component.DrawNoteComponent(1, y, note, screen, textStyle, height, ps.showcw, i == ps.notes.cursor)
		y += height
	}

	page, total := ps.notes.GetCurruntPagePointer()+1, ps.notes.GetTotalPage()+1
	info := fmt.Sprintf("Profile Page %d/%d", page, total)
	if ps.notice != "" {
		info += " · " + ps.notice
	}
	ctx.DrawFooterbar(footerText(info))
}
