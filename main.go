package main

import (
	"termfedi/app"
	"termfedi/config"
	"termfedi/layer"

	"github.com/makachanm/flogger-lib"
)

func main() {
	flogger.Println("Starting termfedi")

	cnf, _, err := config.Load("")
	if err != nil {
		flogger.Errorf("Failed to load config: %v", err)
		panic(err)
	}

	client, err := layer.New(string(cnf.Session.Type), cnf.Session.Url, cnf.Session.Token)
	if err != nil {
		flogger.Errorf("Failed to create api client: %v", err)
		panic(err)
	}

	m := app.NewTerminalScreen()
	err = m.InitTerminalScreen()
	if err != nil {
		flogger.Errorf("Failed to init terminal screen: %v", err)
		panic(err)
	}

	m.RegisterScene("main", app.NewTimelineScreen(cnf, client))
	m.RegisterScene("noti", app.NewNotificationScreen(cnf, client))
	m.RegisterScene("action", app.NewActionScreen(cnf, client))
	m.Mainloop()
}
