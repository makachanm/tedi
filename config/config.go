package config

import "slices"

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

	RefreshOff     = -1
	DefaultRefresh = 30
)

var RefreshSteps = []int{RefreshOff, 10, 20, 30, 60, 120, 300}

const DefaultLimit = 40

var LimitSteps = []int{20, 40, 60, 80, 100, 150, 200}

const DefaultVisibility = "global"

var VisibilitySteps = []string{"global", "quiet", "follower-only", "direct-only"}

const (
	TimelineHome   = "home"
	TimelineLocal  = "local"
	TimelineGlobal = "global"
)

var TimelineSteps = []string{TimelineHome, TimelineLocal, TimelineGlobal}

const (
	TimeRelative = "relative"
	TimeAbsolute = "absolute"
)

var ColorSteps = []string{"white", "silver", "gray", "red", "orange", "yellow", "lime", "green", "aqua", "blue", "fuchsia", "pink"}

type ColorConfiguration struct {
	Text     string `json:"text"`
	Bar      string `json:"bar"`
	Renote   string `json:"renote"`
	Reaction string `json:"reaction"`
	Mention  string `json:"mention"`
}

var DefaultColors = ColorConfiguration{
	Text:     "white",
	Bar:      "white",
	Renote:   "green",
	Reaction: "yellow",
	Mention:  "aqua",
}

type Configuration struct {
	Session SessionConfiguration `json:"session"`
	UI      UIConfiguration      `json:"ui"`
}

type UIConfiguration struct {
	MaxItemHeight int    `json:"maxheight"`
	Layout        string `json:"layout"`
	Refresh       int    `json:"refresh"`
	Limit         int    `json:"limit"`
	Visibility    string `json:"visibility"`
	Timeline      string `json:"timeline"`
	Time          string `json:"time"`
	ShowCW        bool   `json:"showcw"`

	Colors ColorConfiguration `json:"colors"`
}

const (
	LayoutClassic = "classic"
	LayoutFit     = "fit"
)

type SessionConfiguration struct {
	Type  AppType `json:"type"`
	Url   string  `json:"instance_url"`
	Token string  `json:"access_token,omitempty"`

	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`

	Account string `json:"account,omitempty"`

	CredentialID string `json:"credential_id,omitempty"`
}

func (s SessionConfiguration) LoggedIn() bool {
	return s.Type.Valid() && s.Url != "" && s.Token != ""
}

func (s SessionConfiguration) InKeyring() bool {
	return s.CredentialID != ""
}

func (s SessionConfiguration) CredentialLocation() string {
	switch {
	case s.InKeyring():
		return "system keyring"
	case s.Token != "":
		return "config file"
	default:
		return "none"
	}
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
	if c.UI.Layout != LayoutClassic {
		c.UI.Layout = LayoutFit
	}
	if !slices.Contains(RefreshSteps, c.UI.Refresh) {
		c.UI.Refresh = DefaultRefresh
	}
	if !slices.Contains(LimitSteps, c.UI.Limit) {
		c.UI.Limit = DefaultLimit
	}
	if !slices.Contains(VisibilitySteps, c.UI.Visibility) {
		c.UI.Visibility = DefaultVisibility
	}
	if !slices.Contains(TimelineSteps, c.UI.Timeline) {
		c.UI.Timeline = TimelineHome
	}
	if c.UI.Time != TimeAbsolute {
		c.UI.Time = TimeRelative
	}
	colors := &c.UI.Colors
	if !slices.Contains(ColorSteps, colors.Text) {
		colors.Text = DefaultColors.Text
	}
	if !slices.Contains(ColorSteps, colors.Bar) {
		colors.Bar = DefaultColors.Bar
	}
	if !slices.Contains(ColorSteps, colors.Renote) {
		colors.Renote = DefaultColors.Renote
	}
	if !slices.Contains(ColorSteps, colors.Reaction) {
		colors.Reaction = DefaultColors.Reaction
	}
	if !slices.Contains(ColorSteps, colors.Mention) {
		colors.Mention = DefaultColors.Mention
	}
}
