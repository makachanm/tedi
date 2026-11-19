package scene

import (
	"fmt"
	"sync"
	"tedi/app"
	"tedi/app/paging"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type profileView struct {
	userId  string
	profile layer.Profile
	notes   *paging.PagedList[layer.Note]
	loaded  bool
	notice  string
}

type ProfileScreen struct {
	profileView
	lock sync.Mutex

	FetchLayer layer.Client
	detail     *NoteScreen
	config     config.Configuration
	showcw     bool
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
	ps.notes = paging.NewPagedList(ps.config.UI, ps.config.UI.Limit, func(note layer.Note, screenW int) int {
		return component.NoteHeight(note, screenW, 1, ps.showcw)
	})
}

func (ps *ProfileScreen) setCW(show bool) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	ps.showcw = show
	ps.notes.Layout()
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
	ps.notes.Resize(w, max(h-header, 2))
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
	ps.notes.SetItems(notes)
	ps.lock.Unlock()

	ps.resize(screen)
}

func (ps *ProfileScreen) InitScene(screen tcell.Screen, ctx app.ApplicationContext) {
	if !ps.loaded {
		ps.refreshData(screen)
	} else {
		ps.resize(screen)
	}

	screen.Clear()
	ps.drawProfile(screen, ctx)
}

func (ps *ProfileScreen) WindowChangedScene(screen tcell.Screen, ctx app.ApplicationContext) {
	ps.resize(screen)
}

func (ps *ProfileScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx app.ApplicationContext) {
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

		case "tab":
			ctx.Forward("noti", nil)
			return

		case "up":
			ps.lock.Lock()
			ps.notes.MoveCursor(-1)
			ps.lock.Unlock()

		case "down":
			ps.lock.Lock()
			ps.notes.MoveCursor(1)
			ps.lock.Unlock()

		case "enter":
			ps.lock.Lock()
			note, ok := ps.notes.Selected()
			ps.lock.Unlock()
			if ok {
				restore := ps.snapshot()
				ps.detail.open(note)
				ctx.Forward("note", restore)
				return
			}

		case "left", "pgup":
			ps.lock.Lock()
			ps.notes.GoPrev()
			ps.lock.Unlock()

		case "right", "pgdn":
			ps.lock.Lock()
			ps.notes.GoNext()
			ps.lock.Unlock()
		}
	}

	screen.Clear()
	ps.drawProfile(screen, ctx)
}

func (ps *ProfileScreen) drawProfile(screen tcell.Screen, ctx app.ApplicationContext) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	y := 0
	if ps.loaded {
		y = component.DrawProfileHeader(2, 0, ps.profile, screen, textStyle)
	}

	for i, note := range ps.notes.GetCurruntPage() {
		height := ps.notes.Height(note)
		component.DrawNoteComponent(1, y, note, screen, textStyle, height, ps.showcw, i == ps.notes.Cursor)
		y += height
	}

	page, total := ps.notes.GetCurruntPagePointer()+1, ps.notes.GetTotalPage()+1
	info := fmt.Sprintf("Profile Page %d/%d", page, total)
	if ps.notice != "" {
		info += " | " + ps.notice
	}
	ctx.DrawFooterbar(footerText(info, "Back"))
}
