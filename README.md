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
tedi login mastodon.social   # log in
tedi                         # open the client
tedi config                  # settings
tedi status                  # current account
tedi logout                  # log out
tedi help                    # all commands and options
```
