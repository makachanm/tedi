# termfedi
TUI-based Mastodon/Misskey client written in Go.

```
go build .
./termfedi login mastodon.social
./termfedi
```

`login` opens your browser to authorize the app. Over SSH, add `-manual` and paste the code instead.
If you start termfedi without logging in, it asks you to log in first.

`termfedi config` opens the settings screen, where you can change the note height or switch accounts.
`termfedi status` shows the current account, and `termfedi logout` revokes the token.

Settings are saved to `~/.config/termfedi/config.json`.
A `config.json` in the current directory or next to the binary takes precedence.
