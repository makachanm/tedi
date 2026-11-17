package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"termfedi/app"
	"termfedi/auth"
	"termfedi/config"
	"termfedi/layer"

	"golang.org/x/term"
)

const usage = `Usage:
  termfedi [-config FILE]                  start the client (logs in first if needed)
  termfedi login [-manual] [-type T] [-config FILE] [INSTANCE]
                                           log in to a Mastodon or Misskey instance
  termfedi logout [-config FILE]           revoke the token and forget it
  termfedi status [-config FILE]           show which account is configured
  termfedi config [-config FILE]           open the settings screen

Config is read from $TERMFEDI_CONFIG, ./config.json, next to the binary,
or the user config dir (e.g. ~/.config/termfedi/config.json), in that order.
`

func run(args []string) error {
	cmd := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}

	switch cmd {
	case "":
		return runClient(args)
	case "login":
		return runLogin(args)
	case "logout":
		return runLogout(args)
	case "status":
		return runStatus(args)
	case "config":
		return runConfig(args)
	case "help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func newFlagSet(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	configPath := fs.String("config", "", "path to config.json")
	return fs, configPath
}

func interruptContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt)
}

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func warnPermissions(path string) {
	if config.InsecurePermissions(path) {
		fmt.Fprintf(os.Stderr, "warning: %s contains an access token but is readable by other users; run: chmod 600 %s\n", path, path)
	}
}

func login(ctx context.Context, cfg *config.Configuration, path string, opts auth.Options) error {
	opts.In, opts.Out = os.Stdin, os.Stdout
	if opts.DefaultInstance == "" {
		opts.DefaultInstance = cfg.Session.Url
	}

	session, err := auth.Login(ctx, opts)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("login cancelled")
		}
		return fmt.Errorf("login failed: %w", err)
	}

	old := cfg.Session
	if old.LoggedIn() && old.Token != session.Token && old.ClientID != "" {
		auth.Revoke(ctx, old)
	}

	cfg.Session = session
	if err := config.Save(path, *cfg); err != nil {
		return err
	}
	fmt.Printf("Logged in as @%s. Saved to %s\n", session.Account, path)
	return nil
}

func runLogin(args []string) error {
	fs, configPath := newFlagSet("login")
	manual := fs.Bool("manual", false, "paste the authorization code instead of using a local browser callback")
	kind := fs.String("type", "", "server type: mastodon or misskey (auto-detected by default)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return errors.New("login takes at most one instance argument")
	}

	opts := auth.Options{Instance: fs.Arg(0), Manual: *manual}
	if *kind != "" {
		opts.Type = config.AppType(strings.ToLower(*kind))
		if !opts.Type.Valid() {
			return fmt.Errorf("invalid -type %q (expected mastodon or misskey)", *kind)
		}
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	ctx, stop := interruptContext()
	defer stop()
	return login(ctx, &cfg, path, opts)
}

func logout(cfg *config.Configuration, path string) (string, error) {
	ctx, stop := interruptContext()
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var notice string
	switch err := auth.Revoke(ctx, cfg.Session); {
	case err == nil:
		notice = "Token revoked on the server."
	case errors.Is(err, auth.ErrRevokeUnsupported):
		notice = fmt.Sprintf("Remove \"termfedi\" under Settings > Apps on %s to fully revoke access.", cfg.Session.Url)
	default:
		notice = fmt.Sprintf("Could not revoke the token on the server (%v). Revoke it from the authorized apps page on %s.", err, cfg.Session.Url)
	}

	cfg.Session.ClearCredentials()
	if err := config.Save(path, *cfg); err != nil {
		return notice, err
	}
	return notice, nil
}

func runLogout(args []string) error {
	fs, configPath := newFlagSet("logout")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if cfg.Session.Token == "" {
		fmt.Println("Not logged in.")
		return nil
	}

	notice, err := logout(&cfg, path)
	fmt.Println(notice)
	if err != nil {
		return err
	}
	fmt.Printf("Logged out. Removed credentials from %s\n", path)
	return nil
}

func runConfig(args []string) error {
	fs, configPath := newFlagSet("config")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !isInteractive() {
		return errors.New("termfedi config needs a terminal")
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	message := ""
	for {
		updated, action, err := app.RunSettings(cfg, path, message)
		if err != nil {
			return err
		}
		message = ""

		if action == app.SettingsQuit {
			return nil
		}
		if updated != cfg {
			if err := config.Save(path, updated); err != nil {
				return err
			}
			cfg = updated
		}

		switch action {
		case app.SettingsSave:
			fmt.Printf("Saved %s\n", path)
			return nil

		case app.SettingsLogin:
			ctx, stop := interruptContext()
			err := login(ctx, &cfg, path, auth.Options{})
			stop()
			if err != nil {
				message = err.Error()
			} else {
				message = "Logged in as @" + cfg.Session.Account
			}

		case app.SettingsLogout:
			notice, err := logout(&cfg, path)
			if err != nil {
				message = err.Error()
			} else {
				message = "Logged out. " + notice
			}
		}
	}
}

func runStatus(args []string) error {
	fs, configPath := newFlagSet("status")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	fmt.Printf("Config:   %s\n", path)
	if !cfg.Session.LoggedIn() {
		fmt.Println("Status:   not logged in (run `termfedi login`)")
		return nil
	}
	warnPermissions(path)
	fmt.Printf("Instance: %s (%s)\n", cfg.Session.Url, cfg.Session.Type)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	account, err := auth.Verify(ctx, cfg.Session)
	switch {
	case err == nil:
		fmt.Printf("Account:  @%s (token valid)\n", account)
	case errors.Is(err, auth.ErrUnauthorized):
		fmt.Println("Status:   token rejected by the server (run `termfedi login`)")
	default:
		fmt.Printf("Status:   could not reach the server: %v\n", err)
	}
	return nil
}

func ensureSession(cfg *config.Configuration, path string) error {
	ctx, stop := interruptContext()
	defer stop()

	if !cfg.Session.LoggedIn() {
		if !isInteractive() {
			return errors.New("not logged in; run `termfedi login` first")
		}
		fmt.Println("No account configured yet. Let's log in.")
		return login(ctx, cfg, path, auth.Options{})
	}

	warnPermissions(path)

	vctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	account, err := auth.Verify(vctx, cfg.Session)
	switch {
	case err == nil:
		if account != cfg.Session.Account {
			cfg.Session.Account = account
			config.Save(path, *cfg)
		}
		return nil

	case errors.Is(err, auth.ErrUnauthorized):
		if !isInteractive() {
			return errors.New("saved access token was rejected by the server; run `termfedi login`")
		}
		fmt.Printf("The saved access token for %s was rejected (revoked or expired). Please log in again.\n", cfg.Session.Url)
		cfg.Session.ClearCredentials()
		return login(ctx, cfg, path, auth.Options{Instance: cfg.Session.Url, Type: cfg.Session.Type})

	default:
		return nil
	}
}

func runClient(args []string) error {
	fs, configPath := newFlagSet("termfedi")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	cfg, path, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if err := ensureSession(&cfg, path); err != nil {
		return err
	}

	client, err := layer.New(string(cfg.Session.Type), cfg.Session.Url, cfg.Session.Token)
	if err != nil {
		return err
	}

	m := app.NewTerminalScreen()
	if err := m.InitTerminalScreen(); err != nil {
		return err
	}

	m.RegisterScene("main", app.NewTimelineScreen(cfg, client))
	m.RegisterScene("noti", app.NewNotificationScreen(cfg, client))
	m.RegisterScene("action", app.NewActionScreen(cfg, client))
	m.Mainloop()
	return nil
}
