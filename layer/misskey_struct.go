package layer

import "time"

type MisskeyUser struct {
	Id     string  `json:"id"`
	Name   *string `json:"name"`
	Finger string  `json:"username"`
	Host   *string `json:"host"`
}

type MisskeyFile struct {
	Id  string `json:"id"`
	URL string `json:"url"`
}

type MisskeyCreated struct {
	CreatedNote MisskeyNote `json:"createdNote"`
}

type MisskeyNote struct {
	Id             string        `json:"id"`
	CreatedAt      time.Time     `json:"createdAt"`
	Content        *string       `json:"text"`
	Spoilerwarning *string       `json:"cw"`
	User           MisskeyUser   `json:"user"`
	RenoteCount    int           `json:"renoteCount"`
	ReactionCount  int           `json:"reactionCount"`
	Visibility     string        `json:"visibility"`
	Files          []MisskeyFile `json:"files"`
	Renote         *MisskeyNote  `json:"renote,omitempty"`
	ReplyId        *string       `json:"replyId"`
	RepliesCount   int           `json:"repliesCount"`
	Url            *string       `json:"url"`
	Uri            *string       `json:"uri"`
}

type MisskeyNotification struct {
	Id        string       `json:"id"`
	Type      string       `json:"type"`
	CreatedAt time.Time    `json:"createdAt"`
	User      *MisskeyUser `json:"user,omitempty"`
	Note      *MisskeyNote `json:"note,omitempty"`
}

type MisskeyAccount struct {
	MisskeyUser
	Description    *string `json:"description"`
	NotesCount     int     `json:"notesCount"`
	FollowingCount int     `json:"followingCount"`
	FollowersCount int     `json:"followersCount"`
}
