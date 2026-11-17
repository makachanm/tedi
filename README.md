# tedi
TUI-based Mastodon/Misskey client written in Go.

```
go build .
./tedi login mastodon.social
./tedi
```

`tedi login` opens your browser to authorize the app. 

`tedi config` opens the settings screen.

`tedi status` shows the current account.

`tedi logout` revokes the token.

Settings are saved to `~/.config/tedi/config.json`.

## Keys
| Key | Timeline | Notifications | Actions |
|---|---|---|---|
| `q` | quit | quit | back |
| `r` | refresh | refresh | |
| `t` / `1` `2` `3` | next timeline / home, local, global | | |
| `c` | show CW | | |
| `n` | notifications | back | |
| `a` / `Enter` | actions | | do (current mode) |
| `←` `→` / `h` `l` | page | page | select |
| `↑` `↓` / `j` `k` | | | select |
| `f` / `b` | | | react / renote |
| `Tab` | notifications | back | switch mode |
| `Esc` | | back | back |
