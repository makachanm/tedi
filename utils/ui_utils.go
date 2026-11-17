package utils

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func WriteTo(t tcell.Screen, x int, y int, content string, style tcell.Style) {
	curr_x, curr_y := x, y

	for _, cstr := range content {
		var ctr []rune

		width := runewidth.RuneWidth(cstr)
		if width == 0 {
			width = 1
			ctr = []rune{cstr}
			cstr = ' '
		}

		t.SetContent(curr_x, curr_y, cstr, ctr, style)
		curr_x += width
	}
}

func WrapText(text string, width int) []string {
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

func FillLine(s tcell.Screen, y int, text string, style tcell.Style) {
	w, _ := s.Size()
	for x := 0; x < w; x++ {
		s.SetContent(x, y, ' ', nil, style)
	}
	WriteTo(s, 0, y, text, style)
}
