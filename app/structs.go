package app

import "github.com/gdamore/tcell/v2"

type ApplicationScene interface {
	InitScene(screen tcell.Screen, ctx ApplicationContext)
	WindowChangedScene(screen tcell.Screen, ctx ApplicationContext)
	DoScene(screen tcell.Screen, event tcell.Event, ctx ApplicationContext)
}

type ApplicationContext interface {
	Exit(exitcode int)
	Forward(scene string, restore func())
	Back()
	Replace(scene string)
	DrawFooterbar(content string)
	GetCurruntScene() string
}
