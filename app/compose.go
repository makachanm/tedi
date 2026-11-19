package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tedi/component"
	"tedi/config"
	"tedi/layer"

	"github.com/gdamore/tcell/v2"
)

type ComposeScreen struct {
	FetchLayer layer.Client
	config     config.Configuration

	text       string
	visibility layer.Visiblity
	reply      *layer.Note
	notice     string
	confirm    bool
	pending    bool
	scroll     int
	cursor     int
}

func NewComposeScreen(cfg config.Configuration, client layer.Client) *ComposeScreen {
	cs := new(ComposeScreen)
	cs.FetchLayer = client
	cs.config = cfg
	return cs
}

func (cs *ComposeScreen) open(reply *layer.Note) {
	cs.text = ""
	cs.reply = reply
	cs.visibility = layer.VISIBLITY_PUBLIC
	if reply != nil {
		cs.visibility = reply.Visiblity
	}
	cs.notice = ""
	cs.confirm = false
	cs.pending = true
	cs.scroll = 0
	cs.cursor = 0
}

func editorCommand() []string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if args := strings.Fields(os.Getenv(env)); len(args) > 0 {
			return args
		}
	}
	return []string{"vi"}
}

func (cs *ComposeScreen) edit(screen tcell.Screen) {
	dir, err := os.MkdirTemp("", "tedi-note-*")
	if err != nil {
		cs.notice = "Could not create a temp file"
		return
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(file, []byte(cs.text), 0600); err != nil {
		cs.notice = "Could not write the temp file"
		return
	}

	if err := screen.Suspend(); err != nil {
		cs.notice = "Could not leave the screen for the editor"
		return
	}
	args := editorCommand()
	cmd := exec.Command(args[0], append(args[1:], file)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	runErr := cmd.Run()
	screen.Resume()
	screen.Sync()

	if runErr != nil {
		cs.notice = "Editor " + args[0] + " failed"
		return
	}
	d, err := os.ReadFile(file)
	if err != nil {
		cs.notice = "Could not read the temp file"
		return
	}
	cs.text = strings.TrimRight(strings.TrimLeft(string(d), "\r\n"), " \t\r\n")
}

func (cs *ComposeScreen) InitScene(screen tcell.Screen, ctx ApplicationContext) {
	if cs.pending {
		cs.pending = false
		cs.edit(screen)
	}

	screen.Clear()
	cs.draw(screen, ctx)
}

func (cs *ComposeScreen) WindowChangedScene(screen tcell.Screen, ctx ApplicationContext) {
}

var composeItems = []string{"Edit text", "Visibility", "Post", "Discard"}

func visibilityHint(v layer.Visiblity) string {
	switch v {
	case layer.VISIBLITY_PUBLIC:
		return "global: anyone, shown on public timelines."
	case layer.VISIBLITY_QUIET:
		return "quiet: anyone, kept off public timelines."
	case layer.VISIBLITY_FOLLOWER:
		return "follower-only: your followers only."
	default:
		return "direct-only: mentioned users only."
	}
}

func (cs *ComposeScreen) hint() string {
	if composeItems[cs.cursor] == "Visibility" {
		return visibilityHint(cs.visibility)
	}
	return ""
}

func (cs *ComposeScreen) post(ctx ApplicationContext) bool {
	if strings.TrimSpace(cs.text) == "" {
		cs.notice = "Nothing to post"
		return false
	}
	replyTo := ""
	if cs.reply != nil {
		replyTo = cs.reply.Id
	}
	if err := cs.FetchLayer.PostNote(cs.text, cs.visibility, replyTo); err != nil {
		cs.notice = "Post failed"
		return false
	}
	ctx.Back()
	return true
}

func (cs *ComposeScreen) discard(ctx ApplicationContext) bool {
	if strings.TrimSpace(cs.text) != "" && !cs.confirm {
		cs.confirm = true
		cs.notice = "Press again to discard"
		return false
	}
	ctx.Back()
	return true
}

func (cs *ComposeScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		key := keyName(ev)
		discarding := key == "esc" || (key == "enter" && composeItems[cs.cursor] == "Discard")
		if !discarding {
			cs.confirm = false
		}
		cs.notice = ""

		switch key {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			if cs.discard(ctx) {
				return
			}

		case "f1":
			ctx.Forward("compose-help", nil)
			return

		case "up", "k":
			cs.cursor = max(cs.cursor-1, 0)
		case "down", "j":
			cs.cursor = min(cs.cursor+1, len(composeItems)-1)

		case "left", "h":
			if composeItems[cs.cursor] == "Visibility" {
				cs.visibility = (cs.visibility+2)%layer.VISIBLITY_DIRECT + 1
			}
		case "right", "l":
			if composeItems[cs.cursor] == "Visibility" {
				cs.visibility = cs.visibility%layer.VISIBLITY_DIRECT + 1
			}

		case "enter":
			switch composeItems[cs.cursor] {
			case "Edit text":
				cs.edit(screen)
			case "Visibility":
				cs.visibility = cs.visibility%layer.VISIBLITY_DIRECT + 1
			case "Post":
				if cs.post(ctx) {
					return
				}
			case "Discard":
				if cs.discard(ctx) {
					return
				}
			}

		case "v":
			cs.visibility = cs.visibility%layer.VISIBLITY_DIRECT + 1
		case "p":
			if cs.post(ctx) {
				return
			}

		case "pgup":
			cs.scroll -= 5
		case "pgdn", " ":
			cs.scroll += 5
		}
	}

	screen.Clear()
	cs.draw(screen, ctx)
}

func (cs *ComposeScreen) draw(screen tcell.Screen, ctx ApplicationContext) {
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	w, h := screen.Size()

	items := make([]component.MenuItem, len(composeItems))
	for i, label := range composeItems {
		items[i].Label = label
	}
	items[1].Value = strings.ToLower(layer.VisiblityToText(cs.visibility))
	items[1].Choice = true

	y := 1 + component.DrawMenu(1, 1, items, cs.cursor, screen, textStyle)
	y += 1 + component.DrawHint(3, y+1, w-6, cs.hint(), screen, textStyle)
	y++

	author := cs.config.Session.Account
	if author == "" {
		author = "me"
	}
	cs.scroll = component.DrawComposePreview(2, y, w-4, h-1-y, author, cs.text, cs.visibility, cs.reply, screen, textStyle, cs.scroll)

	info := "Compose"
	if cs.reply != nil {
		info = "Reply"
	}
	if cs.notice != "" {
		info += " | " + cs.notice
	}
	ctx.DrawFooterbar(footerText(info, "Back"))
}
