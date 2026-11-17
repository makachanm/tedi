package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"termfedi/config"
)

func NormalizeInstance(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("instance is empty")
	}

	if !strings.Contains(s, "://") && strings.Count(s, "@") >= 1 {
		s = s[strings.LastIndex(s, "@")+1:]
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid instance %q: %w", raw, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid instance %q: no host", raw)
	}

	switch u.Scheme {
	case "https":
	case "http":
		if !isLoopbackHost(u.Hostname()) {
			return "", fmt.Errorf("refusing to log in over plain http to %s; use https", u.Host)
		}
	default:
		return "", fmt.Errorf("invalid instance %q: unsupported scheme %q", raw, u.Scheme)
	}

	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

var misskeyFamily = map[string]bool{
	"misskey": true, "sharkey": true, "firefish": true, "calckey": true,
	"foundkey": true, "cherrypick": true, "iceshrimp": true, "meisskey": true,
}

var mastodonFamily = map[string]bool{
	"mastodon": true, "hometown": true, "glitchsoc": true, "glitch-soc": true,
	"fedibird": true, "kmyblue": true, "pleroma": true, "akkoma": true,
	"gotosocial": true, "iceshrimp.net": true,
}

func DetectType(ctx context.Context, base string) (config.AppType, string, error) {
	software, err := nodeinfoSoftware(ctx, base)
	if err == nil {
		name := strings.ToLower(software)
		switch {
		case misskeyFamily[name]:
			return config.Misskey, software, nil
		case mastodonFamily[name]:
			return config.Mastodon, software, nil
		}
	}

	var instance map[string]any
	if probeErr := getJSON(ctx, base+"/api/v1/instance", "", &instance); probeErr == nil {
		return config.Mastodon, software, nil
	}

	if err != nil {
		return "", "", fmt.Errorf("could not detect server software: %w", err)
	}
	return "", software, fmt.Errorf("unsupported server software %q", software)
}

func nodeinfoSoftware(ctx context.Context, base string) (string, error) {
	var index struct {
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := getJSON(ctx, base+"/.well-known/nodeinfo", "", &index); err != nil {
		return "", err
	}

	var href, bestRel string
	for _, l := range index.Links {
		if strings.HasPrefix(l.Rel, "http://nodeinfo.diaspora.software/ns/schema/") && l.Rel >= bestRel {
			href, bestRel = l.Href, l.Rel
		}
	}
	if href == "" {
		return "", errors.New("no nodeinfo link")
	}
	if u, err := url.Parse(href); err != nil || (u.Scheme != "https" && !isLoopbackHost(u.Hostname())) {
		return "", fmt.Errorf("unexpected nodeinfo url %q", href)
	}

	var info struct {
		Software struct {
			Name string `json:"name"`
		} `json:"software"`
	}
	if err := getJSON(ctx, href, "", &info); err != nil {
		return "", err
	}
	if info.Software.Name == "" {
		return "", errors.New("nodeinfo has no software name")
	}
	return info.Software.Name, nil
}
