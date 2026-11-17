package app

import (
	"fmt"

	"termfedi/config"
	"termfedi/utils"

	"github.com/gdamore/tcell/v2"
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
	itemLayout settingsItem = iota
	itemHeight
	itemFooter
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
	items := []settingsItem{itemLayout}
	if v.cfg.UI.Layout != config.LayoutFit {
		items = append(items, itemHeight)
	}
	items = append(items, itemFooter, itemLogin)
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
	case itemLayout:
		return fmt.Sprintf("%-17s < %s >", "Layout", v.cfg.UI.Layout)
	case itemHeight:
		return fmt.Sprintf("%-17s < %d >", "Note height", v.cfg.UI.MaxItemHeight)
	case itemFooter:
		return fmt.Sprintf("%-17s < %s >", "Footer", v.cfg.UI.Footer)
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

func (v *settingsView) toggle(item settingsItem) {
	switch item {
	case itemLayout:
		if v.cfg.UI.Layout == config.LayoutFit {
			v.cfg.UI.Layout = config.LayoutClassic
		} else {
			v.cfg.UI.Layout = config.LayoutFit
		}
	case itemFooter:
		if v.cfg.UI.Footer == config.FooterInfo {
			v.cfg.UI.Footer = config.FooterGuide
		} else {
			v.cfg.UI.Footer = config.FooterInfo
		}
	}
}

func (v *settingsView) hint(item settingsItem) string {
	switch item {
	case itemLayout:
		if v.cfg.UI.Layout == config.LayoutFit {
			return "fit: each note takes only the lines it needs, with no height limit."
		}
		return "classic: every note takes the same fixed height."
	case itemFooter:
		if v.cfg.UI.Footer == config.FooterInfo {
			return "info: the bottom bar shows only the page and mode."
		}
		return "guide: the bottom bar also lists the key bindings."
	}
	return ""
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
		switch selected {
		case itemLayout, itemFooter:
			v.toggle(selected)
		case itemHeight:
			v.adjustHeight(-1)
		}
	case ev.Key() == tcell.KeyRight || (ev.Key() == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '=')):
		switch selected {
		case itemLayout, itemFooter:
			v.toggle(selected)
		case itemHeight:
			v.adjustHeight(1)
		}

	case ev.Key() == tcell.KeyEnter:
		switch selected {
		case itemLayout, itemFooter:
			v.toggle(selected)
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

func (v *settingsView) draw(s tcell.Screen) {
	s.Clear()
	_, h := s.Size()

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

	y := 1
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

	text, style := v.message, settingsNormal
	if text == "" {
		text, style = v.hint(v.items()[v.cursor]), settingsDim
	}
	w, _ := s.Size()
	for i, line := range utils.WrapText(text, w-6) {
		utils.WriteTo(s, 3, y+1+i, line, style)
	}

	footer := " [Up/Down] Move | [Left/Right] Change | [Enter] Select | [Esc] Quit"
	if v.dirty() {
		footer = " (modified) |" + footer
	}
	utils.FillLine(s, h-1, footer, settingsBar)
	s.Show()
}
