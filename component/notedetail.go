package component

import (
	"strings"

	"tedi/layer"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

func DrawNoteDetail(x int, y int, width int, rows int, note layer.Note, ctx tcell.Screen, style tcell.Style, scroll int) (int, int) {
	if width < 10 || rows <= 0 {
		return 0, 0
	}
	lines := noteDetailLines(note, width, style)

	maxScroll := max(len(lines)-rows, 0)
	scroll = min(max(scroll, 0), maxScroll)

	for i, l := range lines[scroll:min(scroll+rows, len(lines))] {
		utils.WriteTo(ctx, x, y+i, utils.Truncate(l.text, width), l.style)
	}
	return scroll, len(lines)
}

func noteDetailLines(note layer.Note, width int, style tcell.Style) []line {
	var lines []line
	dim := dimStyle(style)

	if note.RenotedBy != "" {
		lines = append(lines, line{"↻ " + note.RenotedBy + " renoted", style.Foreground(tcell.ColorGreen)})
	}

	name := note.Author_name
	if name == "" {
		name = note.Author_finger
	}
	lines = append(lines, line{name, style.Bold(true)})

	meta := []string{"@" + note.Author_finger}
	if !note.Time.IsZero() {
		meta = append(meta, note.Time.Local().Format("2006-01-02 15:04"))
	}
	meta = append(meta, strings.ToLower(layer.VisiblityToText(note.Visiblity)))
	lines = append(lines, wrapLines(strings.Join(meta, " · "), width, dim)...)
	lines = append(lines, line{"", style})

	if note.Spoiler != "" {
		lines = append(lines, wrapLines("CW: "+note.Spoiler, width, style.Foreground(tcell.ColorYellow))...)
		lines = append(lines, line{"", style})
	}

	content := htmlToText(note.Content, true)
	if content == "" && len(note.Medias) == 0 {
		lines = append(lines, line{"(empty)", dim})
	}
	lines = append(lines, wrapLines(content, width, style)...)

	if len(note.Medias) > 0 {
		lines = append(lines, line{"", style})
		for _, media := range note.Medias {
			lines = append(lines, line{"[media]", dim})
			lines = append(lines, wrapLines(media, width, dim)...)
		}
	}

	lines = append(lines, line{"", style})
	lines = append(lines, line{plural(note.RenoteCount, "renote") + " · " + plural(note.ReactionCount, "fav"), dim})
	return lines
}
