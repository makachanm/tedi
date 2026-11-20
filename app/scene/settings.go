package scene

import (
	"fmt"
	"slices"

	"tedi/config"
	"tedi/utils"

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
	itemRefresh
	itemLimit
	itemVisibility
	itemTimeline
	itemTime
	itemCW
	itemTextColor
	itemBarColor
	itemRenoteColor
	itemReactionColor
	itemMentionColor
	itemDefaultColors
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
	items = append(items, itemRefresh, itemLimit, itemVisibility, itemTimeline, itemTime, itemCW,
		itemTextColor, itemBarColor, itemRenoteColor, itemReactionColor, itemMentionColor, itemDefaultColors, itemLogin)
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
	case itemRefresh:
		return fmt.Sprintf("%-17s < %s >", "Auto refresh", refreshText(v.cfg.UI.Refresh))
	case itemLimit:
		return fmt.Sprintf("%-17s < %d >", "Items per load", v.cfg.UI.Limit)
	case itemVisibility:
		return fmt.Sprintf("%-17s < %s >", "Note visibility", v.cfg.UI.Visibility)
	case itemTimeline:
		return fmt.Sprintf("%-17s < %s >", "Start timeline", v.cfg.UI.Timeline)
	case itemTime:
		return fmt.Sprintf("%-17s < %s >", "Time", v.cfg.UI.Time)
	case itemCW:
		return fmt.Sprintf("%-17s < %s >", "Open CW", onOff(v.cfg.UI.ShowCW))
	case itemTextColor, itemBarColor, itemRenoteColor, itemReactionColor, itemMentionColor:
		return fmt.Sprintf("%-17s < %s >", colorLabel(item), *v.color(item))
	case itemDefaultColors:
		return "Reset colors"
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

func (v *settingsView) adjust(item settingsItem, delta int) {
	switch item {
	case itemHeight:
		h := v.cfg.UI.MaxItemHeight + delta
		if h >= config.ItemHeightMin && h <= config.ItemHeightMax {
			v.cfg.UI.MaxItemHeight = h
		}
	case itemRefresh:
		step(config.RefreshSteps, &v.cfg.UI.Refresh, delta)
	case itemLimit:
		step(config.LimitSteps, &v.cfg.UI.Limit, delta)
	case itemVisibility:
		step(config.VisibilitySteps, &v.cfg.UI.Visibility, delta)
	case itemTimeline:
		step(config.TimelineSteps, &v.cfg.UI.Timeline, delta)
	case itemTextColor, itemBarColor, itemRenoteColor, itemReactionColor, itemMentionColor:
		step(config.ColorSteps, v.color(item), delta)
	case itemLayout, itemTime, itemCW:
		v.toggle(item)
	}
}

func refreshText(sec int) string {
	if sec == config.RefreshOff {
		return "off"
	}
	if sec%60 == 0 {
		return fmt.Sprintf("%dm", sec/60)
	}
	return fmt.Sprintf("%ds", sec)
}

func colorLabel(item settingsItem) string {
	switch item {
	case itemTextColor:
		return "Text color"
	case itemBarColor:
		return "Bar color"
	case itemRenoteColor:
		return "Renote color"
	case itemReactionColor:
		return "Reaction color"
	case itemMentionColor:
		return "Mention color"
	}
	return ""
}

func (v *settingsView) color(item settingsItem) *string {
	switch item {
	case itemTextColor:
		return &v.cfg.UI.Colors.Text
	case itemBarColor:
		return &v.cfg.UI.Colors.Bar
	case itemRenoteColor:
		return &v.cfg.UI.Colors.Renote
	case itemReactionColor:
		return &v.cfg.UI.Colors.Reaction
	case itemMentionColor:
		return &v.cfg.UI.Colors.Mention
	}
	return nil
}

func step[T comparable](steps []T, value *T, delta int) {
	i := slices.Index(steps, *value) + delta
	if i < 0 || i >= len(steps) {
		return
	}
	*value = steps[i]
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (v *settingsView) toggle(item settingsItem) {
	switch item {
	case itemLayout:
		if v.cfg.UI.Layout == config.LayoutFit {
			v.cfg.UI.Layout = config.LayoutClassic
		} else {
			v.cfg.UI.Layout = config.LayoutFit
		}
	case itemTime:
		if v.cfg.UI.Time == config.TimeAbsolute {
			v.cfg.UI.Time = config.TimeRelative
		} else {
			v.cfg.UI.Time = config.TimeAbsolute
		}
	case itemCW:
		v.cfg.UI.ShowCW = !v.cfg.UI.ShowCW
	}
}

func (v *settingsView) hint(item settingsItem) string {
	switch item {
	case itemLayout:
		if v.cfg.UI.Layout == config.LayoutFit {
			return "fit: each note takes only the lines it needs, with no height limit."
		}
		return "classic: every note takes the same fixed height."
	case itemRefresh:
		if v.cfg.UI.Refresh == config.RefreshOff {
			return "Timeline and notifications refresh only when you press q."
		}
		return "Timeline and notifications reload every " + refreshText(v.cfg.UI.Refresh) + "."
	case itemLimit:
		return fmt.Sprintf("Load up to %d notes or notifications at a time. Larger numbers take more requests.", v.cfg.UI.Limit)
	case itemVisibility:
		return visibilityHint(configVisibility(v.cfg.UI.Visibility))
	case itemTimeline:
		return "Timeline shown when tedi starts."
	case itemTime:
		if v.cfg.UI.Time == config.TimeAbsolute {
			return "absolute: date and time, like 11-20 14:03."
		}
		return "relative: time since posting, like 5m or 2h."
	case itemTextColor:
		return "Color of notes, menus and other text."
	case itemBarColor:
		return "Color of the bar at the bottom."
	case itemRenoteColor:
		return "Color of renote marks and renote notifications."
	case itemReactionColor:
		return "Color of reaction notifications."
	case itemMentionColor:
		return "Color of mention notifications and reply marks."
	case itemDefaultColors:
		return "Put every color back to the default."
	case itemCW:
		if v.cfg.UI.ShowCW {
			return "Notes with a content warning open by default. z still toggles."
		}
		return "Notes with a content warning stay folded until you press z."
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
		v.adjust(selected, -1)
	case ev.Key() == tcell.KeyRight || (ev.Key() == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '=')):
		v.adjust(selected, 1)

	case ev.Key() == tcell.KeyEnter:
		switch selected {
		case itemLayout, itemTime, itemCW:
			v.toggle(selected)
		case itemDefaultColors:
			v.cfg.UI.Colors = config.DefaultColors
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
	rows := [][2]string{{"Account", account}, {"Server", server}}
	if v.cfg.Session.LoggedIn() {
		rows = append(rows, [2]string{"Token", v.cfg.Session.CredentialLocation()})
	}
	rows = append(rows, [2]string{"File", v.path})
	for _, row := range rows {
		utils.WriteTo(s, 3, y, row[0], settingsDim)
		utils.WriteTo(s, 13, y, row[1], settingsNormal)
		y++
	}

	y++
	for i, item := range v.items() {
		x := 1
		if i == v.cursor {
			x += utils.WriteTo(s, x, y, " > "+v.label(item)+" ", settingsSelected)
		} else {
			x += utils.WriteTo(s, x, y, "   "+v.label(item)+" ", settingsNormal)
		}
		if c := v.color(item); c != nil {
			utils.WriteTo(s, x+1, y, "■■■", settingsNormal.Foreground(tcell.GetColor(*c)))
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
