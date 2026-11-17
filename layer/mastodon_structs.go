package layer

type MastodonUser struct {
	Name   string `json:"display_name"`
	Finger string `json:"acct"`
}

type MastodonMedia struct {
	URL string `json:"url"`
}

type MastodonNote struct {
	Id              string          `json:"id"`
	Content         string          `json:"content"`
	Spoilerwarning  string          `json:"spoiler_text"`
	User            MastodonUser    `json:"account"`
	ReblogsCount    int             `json:"reblogs_count"`
	FavouritesCount int             `json:"favourites_count"`
	Visibility      string          `json:"visibility"`
	Medias          []MastodonMedia `json:"media_attachments"`
	RenotedField    *MastodonNote   `json:"reblog,omitempty"`
}

type MastodonNotification struct {
	Id      string           `json:"id"`
	Type    string           `json:"type"`
	Account MastodonNotiUser `json:"account"`
	Status  *MastodonStatus  `json:"status,omitempty"`
}

type MastodonNotiUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Acct     string `json:"acct"`
}

type MastodonStatus struct {
	Contents string `json:"content"`
}
