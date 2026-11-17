package config

type AppType string

const (
	Mastodon AppType = "mastodon"
	Misskey  AppType = "misskey"
)

func (t AppType) Valid() bool {
	return t == Mastodon || t == Misskey
}

const (
	DefaultMaxItemHeight = 8
	ItemHeightMin        = 5
	ItemHeightMax        = 20
)

type Configuration struct {
	Session SessionConfiguration `json:"session"`
	UI      UIConfiguration      `json:"ui"`
}

type UIConfiguration struct {
	MaxItemHeight int    `json:"maxheight"`
	Layout        string `json:"layout"`
	Footer        string `json:"footer"`
}

const (
	LayoutClassic = "classic"
	LayoutFit     = "fit"

	FooterGuide = "guide"
	FooterInfo  = "info"
)

type SessionConfiguration struct {
	Type  AppType `json:"type"`
	Url   string  `json:"instance_url"`
	Token string  `json:"access_token"`

	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`

	Account string `json:"account,omitempty"`
}

func (s SessionConfiguration) LoggedIn() bool {
	return s.Type.Valid() && s.Url != "" && s.Token != ""
}

func (s *SessionConfiguration) ClearCredentials() {
	s.Token = ""
	s.ClientID = ""
	s.ClientSecret = ""
	s.Account = ""
}

func (c *Configuration) applyDefaults() {
	if c.UI.MaxItemHeight < ItemHeightMin {
		c.UI.MaxItemHeight = DefaultMaxItemHeight
	}
	if c.UI.Layout != LayoutFit {
		c.UI.Layout = LayoutClassic
	}
	if c.UI.Footer != FooterInfo {
		c.UI.Footer = FooterGuide
	}
}
