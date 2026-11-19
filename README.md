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
| Key | |
|---|---|
| `↑` `↓` | select, scroll |
| `←` `→` | page, change value |
| `Enter` | open, run |
| `Esc` / `Backspace` | back, menu on the timeline |
| `Tab` | notifications |
| `F1` | keys for the current screen |

Shortcuts: `x` react, `c` renote, `n` write, `e` profile, `q` refresh, `z` CW, `F2` `F3` `F4` home / local / global.

Writing uses `$VISUAL` or `$EDITOR` (default `vi`).
