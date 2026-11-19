package layer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const misskeyPage = 100
const misskeyThreadRequests = 20

const misskeyDefaultReaction = "❤"

type MisskeyFetch struct {
	token        string
	instance_url *url.URL
	limit        int
}

func NewMisskeyFetch(token string, insurl string, limit int) (*MisskeyFetch, error) {

	u, err := url.Parse(insurl)
	if err != nil {
		return nil, fmt.Errorf("invalid instance url %q: %w", insurl, err)
	}
	return &MisskeyFetch{token: token, instance_url: u, limit: limit}, nil
}

func (m *MisskeyFetch) request(path string, body any) ([]byte, error) {
	u := m.instance_url.JoinPath(path)

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
	if n.Renote != nil && n.Content == nil {
		note := n.Renote.toNote()
		note.RenotedBy = n.User.displayName()
		return note
	}

	note := Note{
		Id:            n.Id,
		Author_id:     n.User.Id,
		Author_finger: n.User.finger(),
		Author_name:   n.User.displayName(),
		Time:          n.CreatedAt,
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
	if n.ReplyId != nil {
		note.ReplyTo = *n.ReplyId
	}

	for _, file := range n.Files {
		note.Medias = append(note.Medias, file.URL)
	}

	return note
}

func (m *MisskeyFetch) timeline(path string, body map[string]any) []Note {
	var notes []Note
	for len(notes) < m.limit {
		page := min(m.limit-len(notes), misskeyPage)
		body["limit"] = page

		d, err := m.request(path, body)
		if err != nil {
			break
		}

		var mnotes []MisskeyNote
		if err := json.Unmarshal(d, &mnotes); err != nil {
			break
		}
		for i := range mnotes {
			notes = append(notes, mnotes[i].toNote())
		}
		if len(mnotes) < page {
			break
		}
		body["untilId"] = mnotes[len(mnotes)-1].Id
	}
	return notes
}

func (m *MisskeyFetch) GetGlobalTimeline() []Note {
	return m.timeline("api/notes/global-timeline", map[string]any{})
}

func (m *MisskeyFetch) GetLocalTimeline() []Note {
	return m.timeline("api/notes/local-timeline", map[string]any{})
}

func (m *MisskeyFetch) GetHomeTimeline() []Note {
	return m.timeline("api/notes/timeline", map[string]any{})
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
	var mnotis []MisskeyNotification
	body := map[string]any{}
	for len(mnotis) < m.limit {
		page := min(m.limit-len(mnotis), misskeyPage)
		body["limit"] = page

		d, err := m.request("api/i/notifications", body)
		if err != nil {
			break
		}

		var got []MisskeyNotification
		if err := json.Unmarshal(d, &got); err != nil {
			break
		}
		mnotis = append(mnotis, got...)
		if len(got) < page {
			break
		}
		body["untilId"] = got[len(got)-1].Id
	}

	notis := make([]Notification, 0, len(mnotis))
	for _, mn := range mnotis {
		n_type := misskeyNotiTypeToValue(mn.Type)
		if n_type == NOTI_UNSUPPORTED {
			continue
		}

		noti := Notification{Id: mn.Id, Type: n_type, Time: mn.CreatedAt}
		if mn.User != nil {
			noti.ReactedUser = User{
				Id:          mn.User.Id,
				User_name:   mn.User.displayName(),
				User_finger: mn.User.finger(),
			}
		}
		if mn.Note != nil {
			noti.NoteId = mn.Note.Id
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
	_, err := m.request("api/notes/create", map[string]any{"renoteId": note_id})
	return err == nil
}

func (m *MisskeyFetch) PostReaction(note_id string) bool {
	_, err := m.request("api/notes/reactions/create", map[string]any{"noteId": note_id, "reaction": misskeyDefaultReaction})
	return err == nil
}

func (m *MisskeyFetch) GetNote(note_id string) (Note, error) {
	d, err := m.request("api/notes/show", map[string]any{"noteId": note_id})
	if err != nil {
		return Note{}, err
	}

	var mNote MisskeyNote
	if err := json.Unmarshal(d, &mNote); err != nil {
		return Note{}, fmt.Errorf("decode note: %w", err)
	}
	return mNote.toNote(), nil
}

func (m *MisskeyFetch) GetThread(note_id string) ([]Note, []Note, error) {
	d, err := m.request("api/notes/conversation", map[string]any{"noteId": note_id, "limit": 100})
	if err != nil {
		return nil, nil, err
	}

	var up []MisskeyNote
	if err := json.Unmarshal(d, &up); err != nil {
		return nil, nil, fmt.Errorf("decode thread: %w", err)
	}

	ancestors := make([]Note, 0, len(up))
	for i := len(up) - 1; i >= 0; i-- {
		ancestors = append(ancestors, up[i].toNote())
	}

	var descendants []Note
	queue := []string{note_id}
	for requests := 0; len(queue) > 0 && requests < misskeyThreadRequests; requests++ {
		d, err := m.request("api/notes/replies", map[string]any{"noteId": queue[0], "limit": 100})
		queue = queue[1:]
		if err != nil {
			return ancestors, descendants, err
		}

		var replies []MisskeyNote
		if err := json.Unmarshal(d, &replies); err != nil {
			return ancestors, descendants, fmt.Errorf("decode thread: %w", err)
		}
		for i := range replies {
			descendants = append(descendants, replies[i].toNote())
			if replies[i].RepliesCount > 0 {
				queue = append(queue, replies[i].Id)
			}
		}
	}
	return ancestors, descendants, nil
}

func (m *MisskeyFetch) GetProfile() (Profile, error) {
	return m.account("api/i", map[string]any{})
}

func (m *MisskeyFetch) GetUser(user_id string) (Profile, error) {
	return m.account("api/users/show", map[string]any{"userId": user_id})
}

func (m *MisskeyFetch) account(path string, body map[string]any) (Profile, error) {
	d, err := m.request(path, body)
	if err != nil {
		return Profile{}, err
	}

	var acct MisskeyAccount
	if err := json.Unmarshal(d, &acct); err != nil {
		return Profile{}, fmt.Errorf("decode profile: %w", err)
	}

	finger := acct.finger()
	if acct.Host == nil {
		finger += "@" + m.instance_url.Host
	}
	profile := Profile{
		Id:             acct.Id,
		Name:           acct.displayName(),
		Finger:         finger,
		NotesCount:     acct.NotesCount,
		FollowingCount: acct.FollowingCount,
		FollowersCount: acct.FollowersCount,
	}
	if acct.Description != nil {
		profile.Bio = *acct.Description
	}
	return profile, nil
}

func (m *MisskeyFetch) GetUserNotes(user_id string) []Note {
	return m.timeline("api/users/notes", map[string]any{"userId": user_id})
}

func (m *MisskeyFetch) PostNote(text string, visibility Visiblity, reply_to string) error {
	_, platform := valueToPlatformVisblity(visibility)
	body := map[string]any{"text": text, "visibility": platform}
	if reply_to != "" {
		body["replyId"] = reply_to
	}
	_, err := m.request("api/notes/create", body)
	return err
}
