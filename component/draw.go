package component

import (
	"fmt"
	"strings"
	"time"

	"tedi/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/net/html"
)

type segment struct {
	text  string
	style tcell.Style
}

type line = segment

func cursorSegments(selected bool, style tcell.Style, segs []segment) []segment {
	if !selected {
		return segs
	}
	return append([]segment{{">  ", style.Bold(true)}}, segs...)
}

func dimStyle(style tcell.Style) tcell.Style {
	return style.Foreground(tcell.ColorGray)
}

func clearRect(s tcell.Screen, x, y, w, h int, style tcell.Style) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			s.SetContent(col, row, ' ', nil, style)
		}
	}
}

func writeSegments(s tcell.Screen, x, y, width int, segs []segment) {
	used := 0
	for _, seg := range segs {
		if used >= width {
			return
		}
		text := seg.text
		if used+runewidth.StringWidth(text) > width {
			text = utils.Truncate(text, width-used)
		}
		used += utils.WriteTo(s, x+used, y, text, seg.style)
	}
}

func drawHeader(s tcell.Screen, x, y, width int, segs []segment, right string, style tcell.Style) {
	rw := runewidth.StringWidth(right)
	avail := width
	if right != "" && width > rw+8 {
		avail = width - rw - 2
		utils.WriteTo(s, x+width-rw, y, right, dimStyle(style))
	}
	writeSegments(s, x, y, avail, segs)
}

func drawBody(s tcell.Screen, x, y, rows, width int, lines []line, style tcell.Style) int {
	if rows <= 0 {
		return 0
	}
	if len(lines) > rows {
		lines = append([]line(nil), lines[:rows]...)
		last := &lines[rows-1]
		last.text = runewidth.Truncate(last.text, width-2, "") + " …"
	}
	for i, l := range lines {
		utils.WriteTo(s, x, y+i, "│", dimStyle(style))
		utils.WriteTo(s, x+2, y+i, utils.Truncate(l.text, width), l.style)
	}
	return len(lines)
}

func wrapLines(text string, width int, style tcell.Style) []line {
	var lines []line
	for _, l := range utils.WrapText(text, width) {
		lines = append(lines, line{l, style})
	}
	return lines
}

const (
	spanNormal = iota
	spanInvisible
	spanEllipsis
)

func htmlToText(src string, full bool) string {
	z := html.NewTokenizer(strings.NewReader(src))
	var b strings.Builder
	var spans []int
	hidden := 0

	for {
		switch z.Next() {
		case html.ErrorToken:
			return tidyText(b.String())

		case html.TextToken:
			if hidden == 0 {
				b.Write(z.Text())
			}

		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "br":
				b.WriteString("\n")
			case "span":
				kind := spanNormal
				for hasAttr {
					var key, val []byte
					key, val, hasAttr = z.TagAttr()
					if string(key) != "class" {
						continue
					}
					for _, class := range strings.Fields(string(val)) {
						switch class {
						case "invisible":
							kind = spanInvisible
						case "ellipsis":
							kind = spanEllipsis
						}
					}
				}
				if full {
					kind = spanNormal
				}
				spans = append(spans, kind)
				if kind == spanInvisible {
					hidden++
				}
			}

		case html.EndTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "p":
				b.WriteString("\n\n")
			case "span":
				if len(spans) == 0 {
					break
				}
				kind := spans[len(spans)-1]
				spans = spans[:len(spans)-1]
				switch kind {
				case spanInvisible:
					hidden--
				case spanEllipsis:
					b.WriteString("…")
				}
			}
		}
	}
}

func tidyText(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	s = strings.Join(lines, "\n")
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}

var AbsoluteTime bool

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	if AbsoluteTime {
		if t.Year() == time.Now().Year() {
			return t.Local().Format("01-02 15:04")
		}
		return t.Local().Format("2006-01-02 15:04")
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return t.Local().Format("2006-01-02")
	}
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
