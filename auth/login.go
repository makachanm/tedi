package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"tedi/config"
)

type Options struct {
	Instance        string
	DefaultInstance string
	Type            config.AppType
	Manual          bool

	In  io.Reader
	Out io.Writer
}

const authorizeTimeout = 10 * time.Minute

func Login(ctx context.Context, opts Options) (config.SessionConfiguration, error) {
	p := &prompter{in: bufio.NewReader(opts.In), out: opts.Out}

	base, err := askInstance(ctx, p, opts)
	if err != nil {
		return config.SessionConfiguration{}, err
	}

	kind := opts.Type
	var software string
	if kind == "" {
		p.printf("Checking %s ...\n", base)
		kind, software, err = DetectType(ctx, base)
		if err != nil {
			if ctx.Err() != nil {
				return config.SessionConfiguration{}, ctx.Err()
			}
			p.printf("Could not detect the server type: %v\n", err)
			kind, err = askType(ctx, p)
			if err != nil {
				return config.SessionConfiguration{}, err
			}
		} else {
			p.printf("Detected %s (%s API).\n", software, kind)
		}
	}

	manual := opts.Manual
	if !manual && !canUseLocalBrowser() {
		manual = true
		p.printf("No local browser detected (SSH or no display); using manual mode.\n")
	}

	authCtx, cancel := context.WithTimeout(ctx, authorizeTimeout)
	defer cancel()

	var session config.SessionConfiguration
	switch kind {
	case config.Mastodon:
		session, err = mastodonLogin(authCtx, p, base, software, manual)
	case config.Misskey:
		session, err = misskeyLogin(authCtx, p, base, manual)
	default:
		err = fmt.Errorf("unsupported server type %q", kind)
	}
	if err != nil {
		return config.SessionConfiguration{}, err
	}

	account, err := Verify(ctx, session)
	if err != nil {
		return config.SessionConfiguration{}, fmt.Errorf("login succeeded but the new token does not work: %w", err)
	}
	session.Account = account
	return session, nil
}

func askInstance(ctx context.Context, p *prompter, opts Options) (string, error) {
	if opts.Instance != "" {
		return NormalizeInstance(opts.Instance)
	}

	def := ""
	if opts.DefaultInstance != "" {
		def = strings.TrimPrefix(strings.TrimRight(opts.DefaultInstance, "/"), "https://")
	}
	for {
		raw, err := p.ask(ctx, "Instance (e.g. mastodon.social)", def)
		if err != nil {
			return "", err
		}
		if raw == "" {
			continue
		}
		base, err := NormalizeInstance(raw)
		if err != nil {
			p.printf("%v\n", err)
			continue
		}
		return base, nil
	}
}

func askType(ctx context.Context, p *prompter) (config.AppType, error) {
	for {
		raw, err := p.ask(ctx, "Server type (mastodon/misskey)", "")
		if err != nil {
			return "", err
		}
		t := config.AppType(strings.ToLower(raw))
		if t.Valid() {
			return t, nil
		}
	}
}

func Verify(ctx context.Context, s config.SessionConfiguration) (string, error) {
	switch s.Type {
	case config.Mastodon:
		return mastodonVerify(ctx, s)
	case config.Misskey:
		return misskeyVerify(ctx, s)
	default:
		return "", fmt.Errorf("unsupported server type %q", s.Type)
	}
}

var ErrRevokeUnsupported = errors.New("token revocation is not supported via the API")

func Revoke(ctx context.Context, s config.SessionConfiguration) error {
	switch s.Type {
	case config.Mastodon:
		return mastodonRevoke(ctx, s)
	default:
		return ErrRevokeUnsupported
	}
}
