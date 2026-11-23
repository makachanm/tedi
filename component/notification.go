package component

import (
	"tedi/layer"
	"tedi/utils"

	"github.com/gdamore/tcell/v2"
)

func notificationLabel(t layer.NotificationType) (string, tcell.Color) {
	switch t {
	case layer.NOTI_MENTION:
		return "Mention", utils.MentionColor
	case layer.NOTI_FAVOURITE:
		return "Favourite", utils.ReactionColor
	case layer.NOTI_RENOTE:
		return "Renote", utils.RenoteColor
	case layer.NOTI_FOLLOW:
		return "Follow", tcell.ColorFuchsia
	default:
		return "Notice", tcell.ColorGray
	}
}

func NotificationHeight(noti layer.Notification, screenW int, x int) int {
	width := screenW - x - 2
	if width < 10 || noti.Type == layer.NOTI_FOLLOW || noti.Content == "" {
		return 2
	}
	return len(wrapLines(htmlToText(noti.Content, false), width-2, tcell.StyleDefault)) + 2
}

func DrawNotificationComponent(x int, y int, noti layer.Notification, ctx tcell.Screen, style tcell.Style, maxheight int, selected bool) {
	screenW, _ := ctx.Size()
	width := screenW - x - 2
	clearRect(ctx, x, y, screenW-x, maxheight, style)
	if width < 10 {
		return
	}

	label, color := notificationLabel(noti.Type)
	user := noti.ReactedUser
	header := []segment{{label + "  ", style.Foreground(color).Bold(true)}}
	if user.User_name != "" {
		header = append(header, segment{user.User_name + " ", style.Bold(true)})
	}
	header = append(header, segment{"@" + user.User_finger, dimStyle(style)})
	drawHeader(ctx, x, y, width, cursorSegments(selected, style, header), relativeTime(noti.Time), style)

	if noti.Type == layer.NOTI_FOLLOW || noti.Content == "" {
		return
	}

	bodyStyle := style
	if noti.Type != layer.NOTI_MENTION {
		bodyStyle = dimStyle(style)
	}
	lines := wrapLines(htmlToText(noti.Content, false), width-2, bodyStyle)
	drawBody(ctx, x, y+1, maxheight-2, width-2, lines, style)
}
