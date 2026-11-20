package component

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"tedi/layer"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

func DrawComposePreview(x int, y int, width int, rows int, author string, text string, visibility layer.Visiblity, reply *layer.Note, ctx tcell.Screen, style tcell.Style, scroll int) int {
	if width < 10 || rows <= 0 {
		return 0
	}
	dim := dimStyle(style)

	lines := []line{{strings.Repeat("─", width), dim}}
	if reply != nil {
		lines = append(lines, line{"↩ Replying to @" + reply.Author_finger, style.Foreground(utils.MentionColor)})
		quoted := wrapLines(htmlToText(reply.Content, false), width-2, dim)
		if len(quoted) > 3 {
			quoted = quoted[:3]
			quoted[2].text = utils.Truncate(quoted[2].text, width-4) + " …"
		}
		for _, q := range quoted {
			lines = append(lines, line{"│ " + q.text, dim})
		}
		lines = append(lines, line{"", style})
	}

	lines = append(lines, line{"@" + author, style.Bold(true)})
	lines = append(lines, line{strings.ToLower(layer.VisiblityToText(visibility)), dim})
	lines = append(lines, line{"", style})

	if strings.TrimSpace(text) == "" {
		lines = append(lines, line{"(empty)", dim})
	} else {
		lines = append(lines, wrapLines(text, width, style)...)
	}
	lines = append(lines, line{"", style})
	lines = append(lines, line{fmt.Sprintf("%d characters", utf8.RuneCountInString(text)), dim})

	scroll = min(max(scroll, 0), max(len(lines)-rows, 0))
	for i, l := range lines[scroll:min(scroll+rows, len(lines))] {
		utils.WriteTo(ctx, x, y+i, utils.Truncate(l.text, width), l.style)
	}
	return scroll
}
