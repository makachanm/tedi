package layer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/makachanm/flogger-lib"
)

const mastodonTimelineLimit = "40"

type MastodonFetch struct {
	token        string
	instance_url *url.URL
}

func NewMastodonFetch(token string, insurl string) (*MastodonFetch, error) {
	flogger.Printf("Initializing MastodonFetch for instance %s", insurl)

	u, err := url.Parse(insurl)
	if err != nil {
		return nil, fmt.Errorf("invalid instance url %q: %w", insurl, err)
	}
	return &MastodonFetch{token: token, instance_url: u}, nil
}

func (m *MastodonFetch) request(method string, path string, query url.Values) ([]byte, error) {
	u := m.instance_url.JoinPath(path)
	u.RawQuery = query.Encode()
	flogger.Printf("Mastodon API call: %s %s", method, u.String())

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	return doRequest(req)
}

func (n *MastodonNote) toNote() Note {
	note := Note{
		Id:            n.Id,
		Author_finger: n.User.Finger,
		Author_name:   n.User.Name,
		Content:       n.Content,
		Spoiler:       n.Spoilerwarning,
		RenoteCount:   n.ReblogsCount,
		ReactionCount: n.FavouritesCount,
		Visiblity:     platformVisblityToValue(n.Visibility),
	}

	if n.RenotedField != nil {
		note.IsRenote = true
		note.Renote = "Renote of " + n.RenotedField.User.Name + "'s Note: <br>" + n.RenotedField.Content
	}

	for _, media := range n.Medias {
		note.Medias = append(note.Medias, media.URL)
	}
	note.HasMedia = len(note.Medias) > 0

	return note
}

func (m *MastodonFetch) timeline(name string, path string, query url.Values) []Note {
	d, err := m.request(http.MethodGet, path, query)
	if err != nil {
		flogger.Errorf("Failed to get %s timeline: %v", name, err)
		return nil
	}

	var mNotes []MastodonNote
	if err := json.Unmarshal(d, &mNotes); err != nil {
		flogger.Errorf("Failed to unmarshal Mastodon notes for %s timeline: %v", name, err)
		return nil
	}
	flogger.Printf("Fetched %d notes from %s timeline", len(mNotes), name)

	notes := make([]Note, 0, len(mNotes))
	for i := range mNotes {
		notes = append(notes, mNotes[i].toNote())
	}
	return notes
}

func (m *MastodonFetch) GetGlobalTimeline() []Note {
	return m.timeline("global", "api/v1/timelines/public", url.Values{"limit": {mastodonTimelineLimit}})
}

func (m *MastodonFetch) GetLocalTimeline() []Note {
	return m.timeline("local", "api/v1/timelines/public", url.Values{"local": {"true"}, "limit": {mastodonTimelineLimit}})
}

func (m *MastodonFetch) GetHomeTimeline() []Note {
	return m.timeline("home", "api/v1/timelines/home", url.Values{"limit": {mastodonTimelineLimit}})
}

func (m *MastodonFetch) GetNotifications() []Notification {
	d, err := m.request(http.MethodGet, "api/v1/notifications", nil)
	if err != nil {
		flogger.Errorf("Failed to get notifications: %v", err)
		return nil
	}

	var mnotis []MastodonNotification
	if err := json.Unmarshal(d, &mnotis); err != nil {
		flogger.Errorf("Failed to unmarshal Mastodon notifications: %v", err)
		return nil
	}
	flogger.Printf("Fetched %d notifications", len(mnotis))

	notis := make([]Notification, 0, len(mnotis))
	for _, mn := range mnotis {
		n_type := platformNotiTypeToValue(mn.Type)
		if n_type == NOTI_UNSUPPORTED {
			continue
		}

		noti := Notification{
			Id:   mn.Id,
			Type: n_type,
			ReactedUser: User{
				Id:          mn.Account.Id,
				User_name:   mn.Account.Username,
				User_finger: mn.Account.Acct,
			},
		}
		if mn.Status != nil {
			noti.Content = mn.Status.Contents
		}
		notis = append(notis, noti)
	}

	return notis
}

func (m *MastodonFetch) PostRenote(note_id string) bool {
	flogger.Printf("Posting renote for note %s", note_id)
	_, err := m.request(http.MethodPost, "api/v1/statuses/"+url.PathEscape(note_id)+"/reblog", nil)
	if err != nil {
		flogger.Errorf("Failed to post renote: %v", err)
	}
	return err == nil
}

func (m *MastodonFetch) PostReaction(note_id string) bool {
	flogger.Printf("Posting reaction for note %s", note_id)
	// Mastodon only has favourite
	_, err := m.request(http.MethodPost, "api/v1/statuses/"+url.PathEscape(note_id)+"/favourite", nil)
	if err != nil {
		flogger.Errorf("Failed to post reaction: %v", err)
	}
	return err == nil
}
