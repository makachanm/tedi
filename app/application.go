package app

import (
	"fmt"
	"os"
	"tedi/config"
	utils "tedi/utils"
	"time"

	"github.com/gdamore/tcell/v2"
)

type TerminalMainApp struct {
	screen        tcell.Screen
	scenes        map[string]ApplicationScene
	currunt_scene string
	appctx        *MainAppContexts

	termination_signal chan int
	transision_signal  chan string
}

func NewTerminalScreen() *TerminalMainApp {
	app := new(TerminalMainApp)

	app.currunt_scene = "main"
	app.scenes = make(map[string]ApplicationScene)
	app.termination_signal = make(chan int, 1)
	app.transision_signal = make(chan string, 1)

	app.appctx = NewMainAppCtx(app.termination_signal, app.transision_signal, app.DrawStatusBar, &app.currunt_scene)
	return app
}

func (t *TerminalMainApp) InitTerminalScreen() error {
	var e error
	t.screen, e = tcell.NewScreen()

	if e != nil {
		return fmt.Errorf("Failed to create new screen: %v", e)
	}

	e = t.screen.Init()
	if e != nil {
		return fmt.Errorf("Failed to init screen: %v", e)
	}

	// TODO: make color to customizeable
	color := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	t.screen.SetStyle(color)

	return nil
}

func (t *TerminalMainApp) RegisterScene(name string, scene ApplicationScene) {
	t.scenes[name] = scene
}

// TODO: add global event handling
func (t *TerminalMainApp) DrawStatusBar(input string) {
	_, h := t.screen.Size()
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	utils.FillLine(t.screen, h-1, input, textStyle)
}

func (t *TerminalMainApp) Mainloop() {
	t.scenes[t.currunt_scene].InitScene(t.screen, t.appctx)

	for {
		select {
		case exitCode := <-t.termination_signal:
			close(t.termination_signal)
			close(t.transision_signal)
			t.screen.Fini()
			os.Exit(exitCode)

		case target := <-t.transision_signal:
			if _, ok := t.scenes[target]; ok {
				t.currunt_scene = target
			} else {
			}

			t.screen.Clear()
			t.screen.Show()

			t.scenes[t.currunt_scene].InitScene(t.screen, t.appctx)
			t.screen.Show()
			continue

		default:
		}

		event := t.screen.PollEvent()

		if _, ok := event.(*tcell.EventResize); ok {
			t.scenes[t.currunt_scene].WindowChangedScene(t.screen, t.appctx)
			t.screen.Clear()
			t.screen.Sync()
		}

		t.scenes[t.currunt_scene].DoScene(t.screen, event, t.appctx)
		t.screen.Show()
	}
}

type MainAppContexts struct {
	termination_signal chan int
	transision_signal  chan string
	scene              *string

	FooterFunc func(string)
}

func NewMainAppCtx(termsig chan int, transsig chan string, footerfunc func(string), scene *string) *MainAppContexts {
	Ctx := new(MainAppContexts)

	Ctx.termination_signal = termsig
	Ctx.transision_signal = transsig
	Ctx.FooterFunc = footerfunc
	Ctx.scene = scene

	return Ctx
}

func (m *MainAppContexts) Exit(exitcode int) {
	m.termination_signal <- exitcode
}

func (m *MainAppContexts) TranslateTo(name string) {
	m.transision_signal <- name
}

func refreshInterval(ui config.UIConfiguration) time.Duration {
	return time.Duration(ui.Refresh) * time.Second
}

func footerText(ui config.UIConfiguration, info string, guide string) string {
	if ui.Footer == config.FooterInfo {
		return " " + info
	}
	return " " + info + " | " + guide
}

func (m *MainAppContexts) DrawFooterbar(content string) {
	m.FooterFunc(content)
}

func (m *MainAppContexts) GetCurruntScene() string {
	return *m.scene
}
