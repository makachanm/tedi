package utils

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func WriteTo(t tcell.Screen, x int, y int, content string, style tcell.Style) int {
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
	return curr_x - x
}

func WrapText(text string, width int) []string {
	if text == "" || width <= 0 {
		return nil
	}

	var lines []string
	for _, para := range strings.Split(text, "\n") {
		if strings.TrimSpace(para) == "" {
			lines = append(lines, "")
			continue
		}

		line := ""
		for _, word := range strings.Fields(para) {
			for runewidth.StringWidth(word) > width {
				if line != "" {
					lines = append(lines, line)
					line = ""
				}
				head := runewidth.Truncate(word, width, "")
				if head == "" {
					_, size := utf8.DecodeRuneInString(word)
					head = word[:size]
				}
				lines = append(lines, head)
				word = word[len(head):]
			}

			switch {
			case word == "":
			case line == "":
				line = word
			case runewidth.StringWidth(line+" "+word) <= width:
				line += " " + word
			default:
				lines = append(lines, line)
				line = word
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func Truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return runewidth.Truncate(text, width, "…")
}

func FillLine(s tcell.Screen, y int, text string, style tcell.Style) {
	w, _ := s.Size()
	for x := 0; x < w; x++ {
		s.SetContent(x, y, ' ', nil, style)
	}
	WriteTo(s, 0, y, text, style)
}
