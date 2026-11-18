package layer

import "time"

type MastodonUser struct {
	Id     string `json:"id"`
	Name   string `json:"display_name"`
	Finger string `json:"acct"`
}

type MastodonMedia struct {
	URL string `json:"url"`
}

type MastodonNote struct {
	Id              string          `json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	Content         string          `json:"content"`
	Spoilerwarning  string          `json:"spoiler_text"`
	User            MastodonUser    `json:"account"`
	ReblogsCount    int             `json:"reblogs_count"`
	FavouritesCount int             `json:"favourites_count"`
	Visibility      string          `json:"visibility"`
	Medias          []MastodonMedia `json:"media_attachments"`
	RenotedField    *MastodonNote   `json:"reblog,omitempty"`
	InReplyTo       *string         `json:"in_reply_to_id"`
}

type MastodonContext struct {
	Ancestors   []MastodonNote `json:"ancestors"`
	Descendants []MastodonNote `json:"descendants"`
}

type MastodonNotification struct {
	Id        string           `json:"id"`
	Type      string           `json:"type"`
	CreatedAt time.Time        `json:"created_at"`
	Account   MastodonNotiUser `json:"account"`
	Status    *MastodonStatus  `json:"status,omitempty"`
}

type MastodonNotiUser struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Acct        string `json:"acct"`
}

type MastodonStatus struct {
	Id       string `json:"id"`
	Contents string `json:"content"`
}

type MastodonAccount struct {
	Id             string `json:"id"`
	DisplayName    string `json:"display_name"`
	Acct           string `json:"acct"`
	Note           string `json:"note"`
	StatusesCount  int    `json:"statuses_count"`
	FollowingCount int    `json:"following_count"`
	FollowersCount int    `json:"followers_count"`
}
