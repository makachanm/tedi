package layer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/makachanm/flogger-lib"
)

const misskeyTimelineLimit = 100

const misskeyDefaultReaction = "❤"

type MisskeyFetch struct {
	token        string
	instance_url *url.URL
}

func NewMisskeyFetch(token string, insurl string) (*MisskeyFetch, error) {
	flogger.Printf("Initializing MisskeyFetch for instance %s", insurl)

	u, err := url.Parse(insurl)
	if err != nil {
		return nil, fmt.Errorf("invalid instance url %q: %w", insurl, err)
	}
	return &MisskeyFetch{token: token, instance_url: u}, nil
}

func (m *MisskeyFetch) request(path string, body any) ([]byte, error) {
	u := m.instance_url.JoinPath(path)
	flogger.Printf("Misskey API call: POST %s", u.String())

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.token)

	return doRequest(req)
}

func (u MisskeyUser) finger() string {
	if u.Host != nil && *u.Host != "" {
		return u.Finger + "@" + *u.Host
	}
	return u.Finger
}

func (u MisskeyUser) displayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	return u.Finger
}

func (n *MisskeyNote) toNote() Note {
	note := Note{
		Id:            n.Id,
		Author_finger: n.User.finger(),
		Author_name:   n.User.displayName(),
		RenoteCount:   n.RenoteCount,
		ReactionCount: n.ReactionCount,
		Visiblity:     platformVisblityToValue(n.Visibility),
	}
	if n.Content != nil {
		note.Content = *n.Content
	}
	if n.Spoilerwarning != nil {
		note.Spoiler = *n.Spoilerwarning
	}

	if n.Renote != nil && n.Content == nil {
		note.IsRenote = true
		var text string
		if n.Renote.Content != nil {
			text = *n.Renote.Content
		}
		note.Renote = "Renote of " + n.Renote.User.displayName() + "'s Note: <br>" + text
	}

	for _, file := range n.Files {
		note.Medias = append(note.Medias, file.URL)
	}
	note.HasMedia = len(note.Medias) > 0

	return note
}

func (m *MisskeyFetch) timeline(name string, path string) []Note {
	d, err := m.request(path, map[string]any{"limit": misskeyTimelineLimit})
	if err != nil {
		flogger.Errorf("Failed to get %s timeline: %v", name, err)
		return nil
	}

	var mnotes []MisskeyNote
	if err := json.Unmarshal(d, &mnotes); err != nil {
		flogger.Errorf("Failed to unmarshal Misskey notes for %s timeline: %v", name, err)
		return nil
	}
	flogger.Printf("Fetched %d notes from %s timeline", len(mnotes), name)

	notes := make([]Note, 0, len(mnotes))
	for i := range mnotes {
		notes = append(notes, mnotes[i].toNote())
	}
	return notes
}

func (m *MisskeyFetch) GetGlobalTimeline() []Note {
	return m.timeline("global", "api/notes/global-timeline")
}

func (m *MisskeyFetch) GetLocalTimeline() []Note {
	return m.timeline("local", "api/notes/local-timeline")
}

func (m *MisskeyFetch) GetHomeTimeline() []Note {
	return m.timeline("home", "api/notes/timeline")
}

func misskeyNotiTypeToValue(s string) NotificationType {
	switch s {
	case "mention", "reply", "quote":
		return NOTI_MENTION
	case "renote":
		return NOTI_RENOTE
	case "reaction":
		return NOTI_FAVOURITE
	case "follow":
		return NOTI_FOLLOW
	default:
		return NOTI_UNSUPPORTED
	}
}

func (m *MisskeyFetch) GetNotifications() []Notification {
	d, err := m.request("api/i/notifications", map[string]any{"limit": misskeyTimelineLimit})
	if err != nil {
		flogger.Errorf("Failed to get notifications: %v", err)
		return nil
	}

	var mnotis []MisskeyNotification
	if err := json.Unmarshal(d, &mnotis); err != nil {
		flogger.Errorf("Failed to unmarshal Misskey notifications: %v", err)
		return nil
	}
	flogger.Printf("Fetched %d notifications", len(mnotis))

	notis := make([]Notification, 0, len(mnotis))
	for _, mn := range mnotis {
		n_type := misskeyNotiTypeToValue(mn.Type)
		if n_type == NOTI_UNSUPPORTED {
			continue
		}

		noti := Notification{Id: mn.Id, Type: n_type}
		if mn.User != nil {
			noti.ReactedUser = User{
				Id:          mn.User.Id,
				User_name:   mn.User.displayName(),
				User_finger: mn.User.finger(),
			}
		}
		if mn.Note != nil {
			switch {
			case mn.Note.Content != nil:
				noti.Content = *mn.Note.Content
			case mn.Note.Renote != nil && mn.Note.Renote.Content != nil:
				noti.Content = *mn.Note.Renote.Content
			}
		}
		notis = append(notis, noti)
	}

	return notis
}

func (m *MisskeyFetch) PostRenote(note_id string) bool {
	flogger.Printf("Posting renote for note %s", note_id)
	_, err := m.request("api/notes/create", map[string]any{"renoteId": note_id})
	if err != nil {
		flogger.Errorf("Failed to post renote: %v", err)
	}
	return err == nil
}

func (m *MisskeyFetch) PostReaction(note_id string) bool {
	flogger.Printf("Posting reaction for note %s", note_id)
	_, err := m.request("api/notes/reactions/create", map[string]any{"noteId": note_id, "reaction": misskeyDefaultReaction})
	if err != nil {
		flogger.Errorf("Failed to post reaction: %v", err)
	}
	return err == nil
}
