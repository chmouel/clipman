# Fork of [clipman](https://github.com/yory8/clipman)

This is a fork of the archived clipman project.

I simply added those features:

- `--err-on-no-selection` exit with an exit 1 when no selection is
  made in the picker. This allows me to do something like this with sway:

  ```generic
  bindsym $super+v exec "clipman pick -t wofi --err-on-no-selection && wtype -M ctrl -M shift v"
  ```

  This will choose a text from clipman and paste it (with the help of wtype) in the current window
  unless I press escape in the picker.

- `--min-char` minimum number of characters before storing

Will be happy to add more or accept contributions.

# Clipman

A basic clipboard manager for Wayland, with support for persisting copy buffers after an application exits.

## Bugs

When you experience a clipboard-related bug, try to see if it still happens without clipman running, as it's more likely to be caused by one of our own known issues rather than wl-clipboard.

## Installing

### From source

Requirements:

- a windows manager that uses `wlr-data-control`, like Sway and other wlroots-based WMs.
- wl-clipboard >= 2.0
- a selector: wofi and bemenu are specially supported, but you can use what you want
- notify-send (optional, for desktop notifications)

[Install go](https://golang.org/doc/install), add `$GOPATH/bin` to your path, then run `go get github.com/chmouel/clipman` OR run `go install` inside this folder.

### Distros

A few distros ship with clipman binaries in their official or unofficial repos.

## Usage

Run the binary in your Sway session by adding `exec wl-paste -t text --watch clipman store` (or `exec wl-paste -t text --watch clipman store 1>> PATH/TO/LOGFILE 2>&1 &` to log errors) at the beginning of your config. It is highly recommended that you run clipman with the `--no-persist` option, see Known Issues.

For primary clipboard support, also add `exec wl-paste -p -t text --watch clipman store -P --histpath="~/.local/share/clipman-primary.json"` (note that both the `-p` in wl-paste and the `-P` in clipman are mandatory in this case).

To query the history and select items, run the binary as `clipman pick -t wofi`. You can assign it to a keybinding: `bindsym $mod+h exec clipman pick -t wofi`.
You can pass additional arguments to the selector like this: `clipman pick --tool wofi -T'--prompt=my-prompt -i'` (both `--prompt` and `-i` are flags of wofi).
You can use a custom selector like this: `clipman pick --print0 --tool=CUSTOM --tool-args="fzf --prompt 'pick > ' --bind 'tab:up' --cycle --read0"`.

To remove items from history, `clipman clear -t wofi` and `clipman clear --all`.

To serve the last history item at startup, add `exec clipman restore` to your Sway config.

For more options: `clipman -h`.

## Known Issues

We only support plain text.
By default, we continue serving the last copied item even after its owner has exited. This means that, unless you run with the `--no-persist` option, you'll always immediately lose rich content; for example:

- vim's visual block mode breaks
- copying images in Firefox breaks
- if you copy a bookmark in Firefox, you won't be able to paste it in another bookmark folder
- if you copy formatted text inside Libre Office you'll lose all formatting on paste

Run `clipman store` with the `--no-persist` option if you are affected. Unfortunately, it seems that there is no way to make them play well together.

## Status

Supporting images or fixing the known issues would require a complete rewrite using wlroots directly.
Clipman is considered feature complete and is now in maintanance mode.

## Related software

- [Clipmon](https://git.sr.ht/~whynothugo/clipmon): a demon specialized in keeping the clipboard alive after an application quits; if that's your only reason for using a clipboard manager, it might be a better fit as it supports any filetype (not just text).

## License

GPL v3.0

2019- (C) yory8 <yory8@users.noreply.github.com>
