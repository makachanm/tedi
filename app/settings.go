package app

import (
	"fmt"
	"strings"

	"termfedi/config"
	"termfedi/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type SettingsAction int

const (
	SettingsQuit SettingsAction = iota
	SettingsSave
	SettingsLogin
	SettingsLogout
)

type settingsItem int

const (
	itemHeight settingsItem = iota
	itemLogin
	itemLogout
	itemSave
)

var (
	settingsNormal   = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	settingsDim      = settingsNormal.Foreground(tcell.ColorGray)
	settingsSelected = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	settingsBar      = settingsSelected
)

type settingsView struct {
	cfg     config.Configuration
	saved   config.Configuration
	path    string
	cursor  int
	message string
	confirm bool
}

func RunSettings(cfg config.Configuration, path string, message string) (config.Configuration, SettingsAction, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return cfg, SettingsQuit, err
	}
	if err := screen.Init(); err != nil {
		return cfg, SettingsQuit, err
	}
	defer screen.Fini()
	screen.SetStyle(settingsNormal)

	v := &settingsView{cfg: cfg, saved: cfg, path: path, message: message}
	for {
		v.draw(screen)
		switch ev := screen.PollEvent().(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if action, done := v.handleKey(ev); done {
				return v.cfg, action, nil
			}
		}
	}
}

func (v *settingsView) items() []settingsItem {
	items := []settingsItem{itemHeight, itemLogin}
	if v.cfg.Session.LoggedIn() {
		items = append(items, itemLogout)
	}
	return append(items, itemSave)
}

func (v *settingsView) dirty() bool {
	return v.cfg != v.saved
}

func (v *settingsView) label(item settingsItem) string {
	switch item {
	case itemHeight:
		return fmt.Sprintf("Note height     < %2d >", v.cfg.UI.MaxItemHeight)
	case itemLogin:
		if v.cfg.Session.LoggedIn() {
			return "Log in again"
		}
		return "Log in"
	case itemLogout:
		return "Log out"
	case itemSave:
		return "Save and quit"
	}
	return ""
}

func (v *settingsView) adjustHeight(delta int) {
	h := v.cfg.UI.MaxItemHeight + delta
	if h < config.ItemHeightMin || h > config.ItemHeightMax {
		return
	}
	v.cfg.UI.MaxItemHeight = h
}

func (v *settingsView) handleKey(ev *tcell.EventKey) (SettingsAction, bool) {
	items := v.items()
	selected := items[v.cursor]

	quit := ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC ||
		(ev.Key() == tcell.KeyRune && ev.Rune() == 'q')
	if !quit {
		v.confirm = false
		v.message = ""
	}

	switch {
	case quit:
		if v.dirty() && !v.confirm {
			v.confirm = true
			v.message = "Unsaved changes. Press Esc again to discard them."
			return 0, false
		}
		return SettingsQuit, true

	case ev.Key() == tcell.KeyUp || (ev.Key() == tcell.KeyRune && ev.Rune() == 'k'):
		if v.cursor > 0 {
			v.cursor--
		}
	case ev.Key() == tcell.KeyDown || (ev.Key() == tcell.KeyRune && ev.Rune() == 'j'):
		if v.cursor < len(items)-1 {
			v.cursor++
		}

	case ev.Key() == tcell.KeyLeft || (ev.Key() == tcell.KeyRune && ev.Rune() == '-'):
		if selected == itemHeight {
			v.adjustHeight(-1)
		}
	case ev.Key() == tcell.KeyRight || (ev.Key() == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '=')):
		if selected == itemHeight {
			v.adjustHeight(1)
		}

	case ev.Key() == tcell.KeyEnter:
		switch selected {
		case itemLogin:
			return SettingsLogin, true
		case itemLogout:
			return SettingsLogout, true
		case itemSave:
			return SettingsSave, true
		}
	}

	return 0, false
}

func wrapText(text string, width int) []string {
	if text == "" || width <= 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && runewidth.StringWidth(line+" "+word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line == "" {
			line = word
		} else {
			line += " " + word
		}
	}
	return append(lines, line)
}

func fillLine(s tcell.Screen, y int, text string, style tcell.Style) {
	w, _ := s.Size()
	for x := 0; x < w; x++ {
		s.SetContent(x, y, ' ', nil, style)
	}
	utils.WriteTo(s, 0, y, text, style)
}

func (v *settingsView) draw(s tcell.Screen) {
	s.Clear()
	_, h := s.Size()

	title := " termfedi settings"
	if v.dirty() {
		title += "  (modified)"
	}
	fillLine(s, 0, title, settingsBar)

	account, server := "not logged in", "-"
	if v.cfg.Session.LoggedIn() {
		account = "@" + v.cfg.Session.Account
		if v.cfg.Session.Account == "" {
			account = "logged in"
		}
	}
	if v.cfg.Session.Url != "" {
		server = fmt.Sprintf("%s (%s)", v.cfg.Session.Url, v.cfg.Session.Type)
	}

	y := 2
	for _, row := range [][2]string{{"Account", account}, {"Server", server}, {"File", v.path}} {
		utils.WriteTo(s, 3, y, row[0], settingsDim)
		utils.WriteTo(s, 13, y, row[1], settingsNormal)
		y++
	}

	y++
	for i, item := range v.items() {
		if i == v.cursor {
			utils.WriteTo(s, 1, y, " > "+v.label(item)+" ", settingsSelected)
		} else {
			utils.WriteTo(s, 1, y, "   "+v.label(item), settingsNormal)
		}
		y++
	}

	w, _ := s.Size()
	for i, line := range wrapText(v.message, w-6) {
		utils.WriteTo(s, 3, y+1+i, line, settingsNormal)
	}

	fillLine(s, h-1, " [Up/Down] Move | [Left/Right] Change | [Enter] Select | [Esc] Quit", settingsBar)
	s.Show()
}
