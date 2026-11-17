package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"tedi/config"
)

const misskeyPermissions = "read:account,read:notifications,write:notes,write:reactions"

func misskeyLogin(ctx context.Context, p *prompter, base string, manual bool) (config.SessionConfiguration, error) {
	session := uuidV4()

	var cb *callbackServer
	if !manual {
		var err error
		cb, err = startCallbackServer(func(q map[string]string) (string, error) {
			if q["session"] != session {
				return "session mismatch", nil
			}
			return "", nil
		})
		if err != nil {
			p.printf("Could not start local callback server (%v); falling back to manual confirmation.\n", err)
		} else {
			defer cb.close()
		}
	}

	params := url.Values{
		"name":       {clientName},
		"permission": {misskeyPermissions},
	}
	if cb != nil {
		params.Set("callback", cb.url)
	}
	authURL := base + "/miauth/" + session + "?" + params.Encode()

	if cb != nil {
		p.printf("\nOpen this URL to authorize tedi:\n\n  %s\n\n", authURL)
		if openBrowser(authURL) {
			p.printf("(Opened in your browser.) ")
		}
		p.printf("Waiting for authorization... (Ctrl+C to cancel)\n")
		if _, err := cb.wait(ctx); err != nil {
			return config.SessionConfiguration{}, err
		}
	} else {
		p.printf("\nOpen this URL in a browser and approve access:\n\n  %s\n\n", authURL)
		if _, err := p.ask(ctx, "Press Enter once you have approved", ""); err != nil {
			return config.SessionConfiguration{}, err
		}
	}

	var res struct {
		OK    bool   `json:"ok"`
		Token string `json:"token"`
	}
	if err := postJSON(ctx, base+"/api/miauth/"+session+"/check", "", struct{}{}, &res); err != nil {
		return config.SessionConfiguration{}, fmt.Errorf("check MiAuth session: %w", err)
	}
	if !res.OK || res.Token == "" {
		return config.SessionConfiguration{}, errors.New("access was not approved (or the session expired); run `tedi login` again")
	}

	return config.SessionConfiguration{
		Type:  config.Misskey,
		Url:   base,
		Token: res.Token,
	}, nil
}

func misskeyVerify(ctx context.Context, s config.SessionConfiguration) (string, error) {
	var me struct {
		Username string `json:"username"`
	}
	if err := postJSON(ctx, strings.TrimRight(s.Url, "/")+"/api/i", s.Token, struct{}{}, &me); err != nil {
		return "", err
	}
	return withHost(me.Username, s.Url), nil
}
