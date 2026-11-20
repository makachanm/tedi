package layer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	mastodonTimelinePage     = 40
	mastodonNotificationPage = 80
)

type MastodonFetch struct {
	token        string
	instance_url *url.URL
	limit        int
}

func NewMastodonFetch(token string, insurl string, limit int) (*MastodonFetch, error) {

	u, err := url.Parse(insurl)
	if err != nil {
		return nil, fmt.Errorf("invalid instance url %q: %w", insurl, err)
	}
	return &MastodonFetch{token: token, instance_url: u, limit: limit}, nil
}

func (m *MastodonFetch) request(method string, path string, query url.Values) ([]byte, error) {
	u := m.instance_url.JoinPath(path)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	return doRequest(req)
}

func (u MastodonUser) displayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Finger
}

func (n *MastodonNote) toNote() Note {
	if n.RenotedField != nil {
		note := n.RenotedField.toNote()
		note.RenotedBy = n.User.displayName()
		return note
	}

	note := Note{
		Id:            n.Id,
		Author_id:     n.User.Id,
		Author_finger: n.User.Finger,
		Author_name:   n.User.displayName(),
		Time:          n.CreatedAt,
		Content:       n.Content,
		Spoiler:       n.Spoilerwarning,
		RenoteCount:   n.ReblogsCount,
		ReactionCount: n.FavouritesCount,
		Visiblity:     platformVisblityToValue(n.Visibility),
	}

	if n.InReplyTo != nil {
		note.ReplyTo = *n.InReplyTo
	}
	note.Url = n.Uri
	if n.Url != nil && *n.Url != "" {
		note.Url = *n.Url
	}

	for _, media := range n.Medias {
		note.Medias = append(note.Medias, media.URL)
		note.MediaIds = append(note.MediaIds, media.Id)
	}

	return note
}

func (m *MastodonFetch) timeline(path string, query url.Values) []Note {
	var notes []Note
	for len(notes) < m.limit {
		page := min(m.limit-len(notes), mastodonTimelinePage)
		query.Set("limit", strconv.Itoa(page))

		d, err := m.request(http.MethodGet, path, query)
		if err != nil {
			break
		}

		var mNotes []MastodonNote
		if err := json.Unmarshal(d, &mNotes); err != nil {
			break
		}
		for i := range mNotes {
			notes = append(notes, mNotes[i].toNote())
		}
		if len(mNotes) < page {
			break
		}
		query.Set("max_id", mNotes[len(mNotes)-1].Id)
	}
	return notes
}

func (m *MastodonFetch) GetGlobalTimeline() []Note {
	return m.timeline("api/v1/timelines/public", url.Values{})
}

func (m *MastodonFetch) GetLocalTimeline() []Note {
	return m.timeline("api/v1/timelines/public", url.Values{"local": {"true"}})
}

func (m *MastodonFetch) GetHomeTimeline() []Note {
	return m.timeline("api/v1/timelines/home", url.Values{})
}

func (m *MastodonFetch) GetNotifications() []Notification {
	var mnotis []MastodonNotification
	query := url.Values{}
	for len(mnotis) < m.limit {
		page := min(m.limit-len(mnotis), mastodonNotificationPage)
		query.Set("limit", strconv.Itoa(page))

		d, err := m.request(http.MethodGet, "api/v1/notifications", query)
		if err != nil {
			break
		}

		var got []MastodonNotification
		if err := json.Unmarshal(d, &got); err != nil {
			break
		}
		mnotis = append(mnotis, got...)
		if len(got) < page {
			break
		}
		query.Set("max_id", got[len(got)-1].Id)
	}

	notis := make([]Notification, 0, len(mnotis))
	for _, mn := range mnotis {
		n_type := platformNotiTypeToValue(mn.Type)
		if n_type == NOTI_UNSUPPORTED {
			continue
		}

		noti := Notification{
			Id:   mn.Id,
			Type: n_type,
			Time: mn.CreatedAt,
			ReactedUser: User{
				Id:          mn.Account.Id,
				User_name:   mn.Account.DisplayName,
				User_finger: mn.Account.Acct,
			},
		}
		if mn.Status != nil {
			noti.Content = mn.Status.Contents
			noti.NoteId = mn.Status.Id
		}
		notis = append(notis, noti)
	}

	return notis
}

func (m *MastodonFetch) PostRenote(note_id string) bool {
	_, err := m.request(http.MethodPost, "api/v1/statuses/"+url.PathEscape(note_id)+"/reblog", nil)
	return err == nil
}

func (m *MastodonFetch) PostReaction(note_id string) bool {
	// Mastodon only has favourite
	_, err := m.request(http.MethodPost, "api/v1/statuses/"+url.PathEscape(note_id)+"/favourite", nil)
	return err == nil
}

func (m *MastodonFetch) GetNote(note_id string) (Note, error) {
	d, err := m.request(http.MethodGet, "api/v1/statuses/"+url.PathEscape(note_id), nil)
	if err != nil {
		return Note{}, err
	}

	var mNote MastodonNote
	if err := json.Unmarshal(d, &mNote); err != nil {
		return Note{}, fmt.Errorf("decode note: %w", err)
	}
	return mNote.toNote(), nil
}

func (m *MastodonFetch) GetThread(note_id string) ([]Note, []Note, error) {
	d, err := m.request(http.MethodGet, "api/v1/statuses/"+url.PathEscape(note_id)+"/context", nil)
	if err != nil {
		return nil, nil, err
	}

	var mctx MastodonContext
	if err := json.Unmarshal(d, &mctx); err != nil {
		return nil, nil, fmt.Errorf("decode thread: %w", err)
	}

	ancestors := make([]Note, 0, len(mctx.Ancestors))
	for i := range mctx.Ancestors {
		ancestors = append(ancestors, mctx.Ancestors[i].toNote())
	}
	descendants := make([]Note, 0, len(mctx.Descendants))
	for i := range mctx.Descendants {
		descendants = append(descendants, mctx.Descendants[i].toNote())
	}
	return ancestors, descendants, nil
}

func (m *MastodonFetch) GetProfile() (Profile, error) {
	return m.account("api/v1/accounts/verify_credentials")
}

func (m *MastodonFetch) GetUser(user_id string) (Profile, error) {
	return m.account("api/v1/accounts/" + url.PathEscape(user_id))
}

func (m *MastodonFetch) account(path string) (Profile, error) {
	d, err := m.request(http.MethodGet, path, nil)
	if err != nil {
		return Profile{}, err
	}

	var acct MastodonAccount
	if err := json.Unmarshal(d, &acct); err != nil {
		return Profile{}, fmt.Errorf("decode profile: %w", err)
	}

	finger := acct.Acct
	if !strings.Contains(finger, "@") {
		finger += "@" + m.instance_url.Host
	}
	return Profile{
		Id:             acct.Id,
		Name:           acct.DisplayName,
		Finger:         finger,
		Bio:            acct.Note,
		NotesCount:     acct.StatusesCount,
		FollowingCount: acct.FollowingCount,
		FollowersCount: acct.FollowersCount,
	}, nil
}

func (m *MastodonFetch) GetUserNotes(user_id string) []Note {
	return m.timeline("api/v1/accounts/"+url.PathEscape(user_id)+"/statuses", url.Values{})
}

func (m *MastodonFetch) PostNote(text string, visibility Visiblity, reply_to string, quote *Note) error {
	if quote != nil {
		link := quote.Url
		if link == "" {
			return fmt.Errorf("quoted note has no link")
		}
		text += "\n\nRE: " + link
	}
	platform, _ := valueToPlatformVisblity(visibility)
	form := url.Values{"status": {text}, "visibility": {platform}}
	if reply_to != "" {
		form.Set("in_reply_to_id", reply_to)
	}

	_, err := m.send(http.MethodPost, "api/v1/statuses", form)
	return err
}

func (m *MastodonFetch) send(method string, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest(method, m.instance_url.JoinPath(path).String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+m.token)

	return doRequest(req)
}

func (m *MastodonFetch) GetNoteSource(note_id string) (string, error) {
	d, err := m.request(http.MethodGet, "api/v1/statuses/"+url.PathEscape(note_id)+"/source", nil)
	if err != nil {
		return "", err
	}

	var source MastodonSource
	if err := json.Unmarshal(d, &source); err != nil {
		return "", fmt.Errorf("decode note source: %w", err)
	}
	return source.Text, nil
}

func (m *MastodonFetch) EditNote(note Note, text string) (Note, error) {
	form := url.Values{"status": {text}, "spoiler_text": {note.Spoiler}}
	for _, id := range note.MediaIds {
		form.Add("media_ids[]", id)
	}

	d, err := m.send(http.MethodPut, "api/v1/statuses/"+url.PathEscape(note.Id), form)
	if err != nil {
		return Note{}, err
	}

	var mNote MastodonNote
	if err := json.Unmarshal(d, &mNote); err != nil {
		return Note{}, fmt.Errorf("decode note: %w", err)
	}
	return mNote.toNote(), nil
}

func (m *MastodonFetch) DeleteNote(note_id string) error {
	_, err := m.request(http.MethodDelete, "api/v1/statuses/"+url.PathEscape(note_id), nil)
	return err
}
