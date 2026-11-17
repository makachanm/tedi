package component

import (
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

func notificationLabel(t layer.NotificationType) (string, tcell.Color) {
	switch t {
	case layer.NOTI_MENTION:
		return "Mention", tcell.ColorAqua
	case layer.NOTI_FAVOURITE:
		return "Favourite", tcell.ColorYellow
	case layer.NOTI_RENOTE:
		return "Renote", tcell.ColorGreen
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
	return len(wrapLines(htmlToText(noti.Content), width-2, tcell.StyleDefault)) + 2
}

func DrawNotificationComponent(x int, y int, noti layer.Notification, ctx tcell.Screen, style tcell.Style, maxheight int) {
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
	drawHeader(ctx, x+1, y, width, header, relativeTime(noti.Time), style)

	if noti.Type == layer.NOTI_FOLLOW || noti.Content == "" {
		return
	}

	bodyStyle := style
	if noti.Type != layer.NOTI_MENTION {
		bodyStyle = dimStyle(style)
	}
	lines := wrapLines(htmlToText(noti.Content), width-2, bodyStyle)
	drawBody(ctx, x+1, y+1, maxheight-2, width-2, lines, style)
}
