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
| Key | Timeline | Notifications |
|---|---|---|
| `Esc` | quit | back |
| `n` / `Tab` | notifications | back |
| `x` / `c` | react / renote selected note | |
| `m` | refresh | refresh |
| `1` `2` `3` | home / local / global timeline | |
| `z` | show CW | |
| `↑` `↓` / `j` `k` | select | select |
| `←` `→` / `h` `l` | page | page |
