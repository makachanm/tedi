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
