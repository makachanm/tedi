package layer

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client interface {
	GetGlobalTimeline() []Note
	GetLocalTimeline() []Note
	GetHomeTimeline() []Note

	GetNotifications() []Notification

	PostRenote(note_id string) bool
	PostReaction(note_id string) bool
}

func New(kind string, instanceURL string, token string) (Client, error) {
	switch kind {
	case "mastodon":
		return NewMastodonFetch(token, instanceURL)
	case "misskey":
		return NewMisskeyFetch(token, instanceURL)
	default:
		return nil, fmt.Errorf("unsupported session type %q", kind)
	}
}

const userAgent = "termfedi"

var httpClient = &http.Client{Timeout: 20 * time.Second}

type HTTPError struct {
	Status string
	Code   int
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("API error: %s", e.Status)
}

func doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, &HTTPError{Status: resp.Status, Code: resp.StatusCode, Body: string(body)}
	}
	return body, nil
}
