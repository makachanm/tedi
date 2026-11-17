package component

import (
	"strings"

	"termfedi/layer"
	"termfedi/utils"

	"github.com/gdamore/tcell/v2"
)

func DrawNoteComponent(x int, y int, note layer.Note, ctx tcell.Screen, style tcell.Style, maxheight int, viewcw bool) {
	screenW, _ := ctx.Size()
	width := screenW - x - 2
	clearRect(ctx, x, y, screenW-x, maxheight, style)
	if width < 10 {
		return
	}

	dim := dimStyle(style)
	row := y

	if note.RenotedBy != "" {
		writeSegments(ctx, x+1, row, width, []segment{{"↻ " + note.RenotedBy + " renoted", style.Foreground(tcell.ColorGreen)}})
		row++
	}

	name := note.Author_name
	if name == "" {
		name = note.Author_finger
	}
	drawHeader(ctx, x+1, row, width, []segment{
		{name, style.Bold(true)},
		{" @" + note.Author_finger, dim},
	}, relativeTime(note.Time), style)
	row++

	lastRow := y + maxheight - 2
	metaRow := row + drawBody(ctx, x+1, row, lastRow-row, width-2, noteLines(note, viewcw, width-2, style), style)

	meta := []string{strings.ToLower(layer.VisiblityToText(note.Visiblity))}
	if note.RenoteCount > 0 {
		meta = append(meta, plural(note.RenoteCount, "renote"))
	}
	if note.ReactionCount > 0 {
		meta = append(meta, plural(note.ReactionCount, "fav"))
	}
	writeSegments(ctx, x+1, metaRow, width, []segment{{strings.Join(meta, " · "), dim}})
}

func NoteHeight(note layer.Note, screenW int, x int, viewcw bool) int {
	width := screenW - x - 2
	if width < 10 {
		return 4
	}
	height := len(noteLines(note, viewcw, width-2, tcell.StyleDefault)) + 3
	if note.RenotedBy != "" {
		height++
	}
	return height
}

func noteLines(note layer.Note, viewcw bool, width int, style tcell.Style) []line {
	var lines []line
	dim := dimStyle(style)

	if note.Spoiler != "" {
		lines = append(lines, wrapLines("CW: "+note.Spoiler, width, style.Foreground(tcell.ColorYellow))...)
		if !viewcw {
			hidden := "content hidden"
			if n := len(note.Medias); n > 0 {
				hidden += ", " + plural(n, "attachment")
			}
			return append(lines, line{"(" + hidden + ")", dim})
		}
	}

	lines = append(lines, wrapLines(htmlToText(note.Content), width, style)...)
	for _, media := range note.Medias {
		lines = append(lines, line{utils.Truncate("[media] "+media, width), dim})
	}
	if len(lines) == 0 {
		lines = append(lines, line{"(empty)", dim})
	}
	return lines
}
