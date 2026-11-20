package utils

import "github.com/gdamore/tcell/v2"

var (
	TextColor     = tcell.ColorWhite
	BarColor      = tcell.ColorWhite
	RenoteColor   = tcell.ColorGreen
	ReactionColor = tcell.ColorYellow
	MentionColor  = tcell.ColorAqua
)

func TextStyle() tcell.Style {
	return tcell.StyleDefault.Foreground(TextColor).Background(tcell.ColorBlack)
}

func BarStyle() tcell.Style {
	r, g, b := BarColor.RGB()
	if r*299+g*587+b*114 < 128000 {
		return tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(BarColor)
	}
	return tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(BarColor)
}
