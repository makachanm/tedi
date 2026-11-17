package layer

type MisskeyNoteTransaction struct {
	Limit int `json:"limit"`
}

type MisskeyUser struct {
	Name   string `json:"name"`
	Finger string `json:"username"`
	Host   string `json:"host"`
}

type MisskeyNote struct {
	Id              string      `json:"id"`
	Content         string      `json:"text"`
	Spoilerwarning  *string     `json:"cw"`
	User            MisskeyUser `json:"user"`
	RenoteCount     int         `json:"renoteCount"`
	FavouritesCount int         `json:"reactionCount"`
	Visibility      string      `json:"visibility"`
}

type MisskeyNotification struct {
	Id      string           `json:"id"`
	Type    string           `json:"type"`
	Account MastodonNotiUser `json:"account"`
	Status  MastodonStatus   `json:"status,omitempty"`
}
