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

func (cs *ComposeScreen) DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext) {
	if ev, ok := event.(*tcell.EventKey); ok {
		key := keyName(ev)
		if key != "esc" {
			cs.confirm = false
		}
		cs.notice = ""

		switch key {
		case "C-c":
			ctx.Exit(0)

		case "esc":
			if strings.TrimSpace(cs.text) != "" && !cs.confirm {
				cs.confirm = true
				cs.notice = "Press Esc again to discard this note"
				break
			}
			ctx.Back()
			return

		case "f1":
			ctx.Forward("compose-help", nil)
			return

		case "enter":
			cs.edit(screen)

		case "v":
			cs.visibility = cs.visibility%layer.VISIBLITY_DIRECT + 1

		case "p":
			if strings.TrimSpace(cs.text) == "" {
				cs.notice = "Nothing to post"
				break
			}
			replyTo := ""
			if cs.reply != nil {
				replyTo = cs.reply.Id
			}
			if err := cs.FetchLayer.PostNote(cs.text, cs.visibility, replyTo); err != nil {
				cs.notice = "Post failed"
				break
			}
			ctx.Back()
			return

		case "up", "k":
			cs.scroll--
		case "down", "j":
			cs.scroll++
		}
	}

	screen.Clear()
	cs.draw(screen, ctx)
}

func (cs *ComposeScreen) draw(screen tcell.Screen, ctx ApplicationContext) {
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	w, h := screen.Size()

	author := cs.config.Session.Account
	if author == "" {
		author = "me"
	}
	cs.scroll = component.DrawComposePreview(2, 1, w-4, h-2, author, cs.text, cs.visibility, cs.reply, screen, textStyle, cs.scroll)

	info := "Compose"
	if cs.reply != nil {
		info = "Reply"
	}
	info += " · " + strings.ToLower(layer.VisiblityToText(cs.visibility))
	if cs.notice != "" {
		info += " · " + cs.notice
	}
	ctx.DrawFooterbar(footerText(info))
}
