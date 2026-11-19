package component

import (
	"sort"
	"strings"

	"tedi/layer"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

const threadMaxDepth = 6

func DrawNoteDetail(x int, y int, width int, rows int, ancestors []layer.Note, note layer.Note, replies []layer.Note, ctx tcell.Screen, style tcell.Style, scroll int, jump bool) (int, int) {
	if width < 10 || rows <= 0 {
		return 0, 0
	}
	dim := dimStyle(style)
	rule := line{strings.Repeat("─", width), dim}

	var lines []line
	for _, n := range ancestors {
		lines = append(lines, threadEntryLines(n, 0, width, style)...)
	}
	if len(ancestors) > 0 {
		lines = append(lines, rule)
	}

	focus := len(lines)
	lines = append(lines, noteDetailLines(note, width, style)...)

	if len(replies) > 0 {
		lines = append(lines, line{"", style}, rule)
		ordered, depths := replyTree(note.Id, replies)
		for i, n := range ordered {
			lines = append(lines, threadEntryLines(n, depths[i], width, style)...)
		}
	}

	if jump {
		scroll = focus
	}
	maxScroll := max(len(lines)-rows, 0)
	scroll = min(max(scroll, 0), maxScroll)

	for i, l := range lines[scroll:min(scroll+rows, len(lines))] {
		utils.WriteTo(ctx, x, y+i, utils.Truncate(l.text, width), l.style)
	}
	return scroll, len(lines)
}

func replyTree(rootId string, replies []layer.Note) ([]layer.Note, []int) {
	children := make(map[string][]layer.Note)
	for _, n := range replies {
		children[n.ReplyTo] = append(children[n.ReplyTo], n)
	}
	for _, list := range children {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Time.Before(list[j].Time) })
	}

	var ordered []layer.Note
	var depths []int
	seen := make(map[string]bool)

	var walk func(id string, depth int)
	walk = func(id string, depth int) {
		for _, n := range children[id] {
			if seen[n.Id] {
				continue
			}
			seen[n.Id] = true
			ordered = append(ordered, n)
			depths = append(depths, depth)
			walk(n.Id, depth+1)
		}
	}
	walk(rootId, 0)

	for _, n := range replies {
		if !seen[n.Id] {
			seen[n.Id] = true
			ordered = append(ordered, n)
			depths = append(depths, 0)
		}
	}
	return ordered, depths
}

func threadEntryLines(note layer.Note, depth int, width int, style tcell.Style) []line {
	indent := strings.Repeat("  ", min(depth, threadMaxDepth))
	w := max(width-len(indent), 10)
	dim := dimStyle(style)

	name := note.Author_name
	if name == "" {
		name = note.Author_finger
	}
	header := name + " @" + note.Author_finger
	if t := relativeTime(note.Time); t != "" {
		header += "  " + t
	}

	var lines []line
	lines = append(lines, line{indent + utils.Truncate(header, w), style.Bold(true)})
	if note.Spoiler != "" {
		for _, l := range wrapLines("CW: "+note.Spoiler, w, style.Foreground(tcell.ColorYellow)) {
			lines = append(lines, line{indent + l.text, l.style})
		}
	}
	for _, l := range wrapLines(htmlToText(note.Content, true), w, style) {
		lines = append(lines, line{indent + l.text, l.style})
	}
	for _, media := range note.Medias {
		for _, l := range wrapLines("[media] "+media, w, dim) {
			lines = append(lines, line{indent + l.text, l.style})
		}
	}
	return append(lines, line{"", style})
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
	lines = append(lines, wrapLines(strings.Join(meta, "  "), width, dim)...)
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
	lines = append(lines, line{plural(note.RenoteCount, "renote") + "  " + plural(note.ReactionCount, "fav"), dim})
	return lines
}
