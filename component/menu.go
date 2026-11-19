package component

import (
	"fmt"

	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

type MenuItem struct {
	Label  string
	Value  string
	Choice bool
}

func DrawMenu(x int, y int, items []MenuItem, cursor int, ctx tcell.Screen, style tcell.Style) int {
	selected := style.Reverse(true)
	for i, item := range items {
		text := item.Label
		switch {
		case item.Choice:
			text = fmt.Sprintf("%-17s < %s >", item.Label, item.Value)
		case item.Value != "":
			text = fmt.Sprintf("%-17s (%s)", item.Label, item.Value)
		}
		if i == cursor {
			utils.WriteTo(ctx, x, y+i, " > "+text+" ", selected)
		} else {
			utils.WriteTo(ctx, x, y+i, "   "+text, style)
		}
	}
	return len(items)
}

func DrawActionBar(x int, y int, actions []string, cursor int, ctx tcell.Screen, style tcell.Style) {
	for i, action := range actions {
		label := " " + action + " "
		if i == cursor {
			x += utils.WriteTo(ctx, x, y, label, style.Reverse(true))
		} else {
			x += utils.WriteTo(ctx, x, y, label, dimStyle(style))
		}
		x += 2
	}
}

func DrawHint(x int, y int, width int, text string, ctx tcell.Screen, style tcell.Style) int {
	lines := utils.WrapText(text, width)
	for i, l := range lines {
		utils.WriteTo(ctx, x, y+i, l, dimStyle(style))
	}
	return len(lines)
}
