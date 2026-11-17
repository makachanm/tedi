package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"termfedi/config"
)

const oobRedirect = "urn:ietf:wg:oauth:2.0:oob"

const (
	mastodonGranularScopes = "read write:favourites write:statuses"
	mastodonLegacyScopes   = "read write"
)

var granularScopeSoftware = map[string]bool{
	"mastodon": true, "hometown": true, "glitchsoc": true, "glitch-soc": true,
	"fedibird": true, "kmyblue": true,
}

type mastodonApp struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func mastodonLogin(ctx context.Context, p *prompter, base string, software string, manual bool) (config.SessionConfiguration, error) {
	scopes := mastodonLegacyScopes
	if software == "" || granularScopeSoftware[strings.ToLower(software)] {
		scopes = mastodonGranularScopes
	}

	state := randomToken(16)
	verifier, challenge := pkce()

	var cb *callbackServer
	redirect := oobRedirect
	if !manual {
		var err error
		cb, err = startCallbackServer(func(q map[string]string) (string, error) {
			if q["state"] != state {
				return "state mismatch", nil
			}
			if e := q["error"]; e != "" {
				if d := q["error_description"]; d != "" {
					e += ": " + d
				}
				return "", fmt.Errorf("authorization denied (%s)", e)
			}
			if q["code"] == "" {
				return "missing code", nil
			}
			return "", nil
		})
		if err != nil {
			p.printf("Could not start local callback server (%v); falling back to manual code entry.\n", err)
		} else {
			defer cb.close()
			redirect = cb.url
		}
	}

	var app mastodonApp
	err := postForm(ctx, base+"/api/v1/apps", url.Values{
		"client_name":   {clientName},
		"redirect_uris": {redirect},
		"scopes":        {scopes},
		"website":       {clientSite},
	}, &app)
	if err != nil {
		return config.SessionConfiguration{}, fmt.Errorf("register app on %s: %w", base, err)
	}
	if app.ClientID == "" || app.ClientSecret == "" {
		return config.SessionConfiguration{}, errors.New("register app: server returned no client credentials")
	}

	authURL := base + "/oauth/authorize?" + url.Values{
		"client_id":             {app.ClientID},
		"redirect_uri":          {redirect},
		"response_type":         {"code"},
		"scope":                 {scopes},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()

	var code string
	if cb != nil {
		p.printf("\nOpen this URL to authorize termfedi:\n\n  %s\n\n", authURL)
		if openBrowser(authURL) {
			p.printf("(Opened in your browser.) ")
		}
		p.printf("Waiting for authorization... (Ctrl+C to cancel)\n")

		q, err := cb.wait(ctx)
		if err != nil {
			return config.SessionConfiguration{}, err
		}
		code = q["code"]
	} else {
		p.printf("\nOpen this URL in a browser, approve access, and copy the code it shows:\n\n  %s\n\n", authURL)
		for code == "" {
			code, err = p.ask(ctx, "Authorization code", "")
			if err != nil {
				return config.SessionConfiguration{}, err
			}
		}
	}

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	err = postForm(ctx, base+"/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {app.ClientID},
		"client_secret": {app.ClientSecret},
		"redirect_uri":  {redirect},
		"code_verifier": {verifier},
		"scope":         {scopes},
	}, &tok)
	if err != nil {
		return config.SessionConfiguration{}, fmt.Errorf("exchange authorization code: %w", err)
	}
	if tok.AccessToken == "" {
		return config.SessionConfiguration{}, errors.New("exchange authorization code: server returned no token")
	}

	return config.SessionConfiguration{
		Type:         config.Mastodon,
		Url:          base,
		Token:        tok.AccessToken,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	}, nil
}

func mastodonVerify(ctx context.Context, s config.SessionConfiguration) (string, error) {
	var acct struct {
		Acct string `json:"acct"`
	}
	if err := getJSON(ctx, strings.TrimRight(s.Url, "/")+"/api/v1/accounts/verify_credentials", s.Token, &acct); err != nil {
		return "", err
	}
	return withHost(acct.Acct, s.Url), nil
}

func mastodonRevoke(ctx context.Context, s config.SessionConfiguration) error {
	if s.ClientID == "" || s.ClientSecret == "" {
		return errors.New("no client credentials stored (token was not created by `termfedi login`)")
	}
	return postForm(ctx, strings.TrimRight(s.Url, "/")+"/oauth/revoke", url.Values{
		"client_id":     {s.ClientID},
		"client_secret": {s.ClientSecret},
		"token":         {s.Token},
	}, nil)
}

func withHost(user string, base string) string {
	if user == "" || strings.Contains(user, "@") {
		return user
	}
	if u, err := url.Parse(base); err == nil && u.Host != "" {
		return user + "@" + u.Host
	}
	return user
}
