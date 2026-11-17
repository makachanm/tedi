package layer

import "fmt"

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
