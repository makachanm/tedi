package component

import (
	"fmt"
	"strings"

	"tedi/layer"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

const profileBioLines = 6

func ProfileHeight(profile layer.Profile, screenW int, x int) int {
	return len(profileLines(profile, screenW-x-2, tcell.StyleDefault))
}

func DrawProfileHeader(x int, y int, profile layer.Profile, ctx tcell.Screen, style tcell.Style) int {
	screenW, _ := ctx.Size()
	width := screenW - x - 2
	lines := profileLines(profile, width, style)
	for i, l := range lines {
		utils.WriteTo(ctx, x, y+i, utils.Truncate(l.text, width), l.style)
	}
	return len(lines)
}

func profileLines(profile layer.Profile, width int, style tcell.Style) []line {
	if width < 10 {
		return []line{{"", style}}
	}
	dim := dimStyle(style)

	name := profile.Name
	if name == "" {
		name = profile.Finger
	}
	lines := []line{{name, style.Bold(true)}}
	if profile.Finger != "" {
		lines = append(lines, line{"@" + profile.Finger, dim})
	}

	bio := wrapLines(htmlToText(profile.Bio, true), width, style)
	if len(bio) > profileBioLines {
		bio = bio[:profileBioLines]
		bio[profileBioLines-1].text = utils.Truncate(bio[profileBioLines-1].text, width-2) + " …"
	}
	if len(bio) > 0 {
		lines = append(lines, line{"", style})
		lines = append(lines, bio...)
	}

	lines = append(lines, line{"", style})
	stats := []string{
		plural(profile.NotesCount, "note"),
		fmt.Sprintf("%d following", profile.FollowingCount),
		plural(profile.FollowersCount, "follower"),
	}
	lines = append(lines, line{strings.Join(stats, " · "), dim})
	return append(lines, line{strings.Repeat("─", width), dim})
}
