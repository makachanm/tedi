# tedi
TUI-based Mastodon/Misskey client written in Go.

## Install
```
curl -fsSL https://raw.githubusercontent.com/makachanm/tedi/main/install.sh | sh
```
`.deb`, `.rpm` and tarballs for amd64/arm64 are on the [latest release](https://github.com/makachanm/tedi/releases/tag/latest), rebuilt on every push to main.

To build from source:
```
go build .
```

## Usage
```
tedi login mastodon.social
tedi
```

`tedi login` opens your browser to authorize the app. 

`tedi config` opens the settings screen.

`tedi status` shows the current account.

`tedi logout` revokes the token.

Settings are saved to `~/.config/tedi/config.json`.
The access token is kept in the system keyring (Secret Service, Keychain, Credential Manager) when one is available, and in the config file (mode 600) otherwise.

## Keys
| Key | Timeline | Notifications | Profile | Note |
|---|---|---|---|---|
| `Esc` | quit | back | back | back |
| `Enter` | open note | open note / user | open note | author's profile |
| `w` | notifications | | | |
| `e` | my profile | | | |
| `x` / `c` | react / renote | | react / renote | react / renote |
| `q` | refresh | refresh | refresh | |
| `F2` `F3` `F4` | home / local / global timeline | | | |
| `z` | show CW | | show CW | |
| `↑` `↓` / `j` `k` | select | select | select | scroll |
| `←` `→` / `h` `l` | page | page | page | page |
| `F1` | key guide | key guide | key guide | key guide |
