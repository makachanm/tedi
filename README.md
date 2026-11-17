# termfedi
TUI-based Mastodon/Misskey client written in Go.

```
go build .
./termfedi login mastodon.social
./termfedi
```

`termfedi login` opens your browser to authorize the app. 

`termfedi config` opens the settings screen.

`termfedi status` shows the current account.

`termfedi logout` revokes the token.

Settings are saved to `~/.config/termfedi/config.json`.
