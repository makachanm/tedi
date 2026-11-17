package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	clientName = "termfedi"
	clientSite = "https://github.com/makachanm/termfedi"
	userAgent  = "termfedi"
)

var ErrUnauthorized = errors.New("access token was rejected by the server")

var httpClient = &http.Client{Timeout: 20 * time.Second}

type statusError struct {
	code   int
	status string
	body   string
}

func (e *statusError) Error() string {
	msg := strings.TrimSpace(e.body)
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	if msg == "" {
		return "server returned " + e.status
	}
	return fmt.Sprintf("server returned %s: %s", e.status, msg)
}

func (e *statusError) Unwrap() error {
	if e.code == http.StatusUnauthorized || e.code == http.StatusForbidden {
		return ErrUnauthorized
	}
	return nil
}

func do(req *http.Request, out any) error {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &statusError{code: resp.StatusCode, status: resp.Status, body: string(body)}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("unexpected response from %s: %w", req.URL.Host, err)
	}
	return nil
}

func getJSON(ctx context.Context, rawURL string, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return do(req, out)
}

func postForm(ctx context.Context, rawURL string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return do(req, out)
}

func postJSON(ctx context.Context, rawURL string, token string, in any, out any) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return do(req, out)
}
