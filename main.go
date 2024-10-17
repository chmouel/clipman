// GPL v3.0
// 2019- (C) yory8 <yory8@users.noreply.github.com>
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
)

const version = "1.6.2"

var (
	app      = kingpin.New("clipman", "A clipboard manager for Wayland")
	histpath = app.Flag("histpath", "Path of history file").Default("~/.local/share/clipman.json").String()
	alert    = app.Flag("notify", "Send desktop notifications on errors").Bool()
	primary  = app.Flag("primary", "Serve item to the primary clipboard").Default("false").Bool()

	storer    = app.Command("store", "Record clipboard events (run as argument to `wl-paste --watch`)")
	maxDemon  = storer.Flag("max-items", "history size").Default("15").Int()
	noPersist = storer.Flag("no-persist", "Don't persist a copy buffer after a program exits").Short('P').Default("false").Bool()
	minChar   = storer.Flag("min-char", "Minimum number of characters before storing").Default("-1").Int()
	unix      = storer.Flag("unix", "Normalize line endings to LF").Bool()

	picker             = app.Command("pick", "Pick an item from clipboard history")
	maxPicker          = picker.Flag("max-items", "scrollview length").Default("15").Int()
	pickTool           = picker.Flag("tool", "Which selector to use: wofi/bemenu/CUSTOM/dmenu/rofi/STDOUT").Short('t').Required().String()
	pickToolArgs       = picker.Flag("tool-args", "Extra arguments to pass to the --tool").Short('T').Default("").String()
	pickEsc            = picker.Flag("print0", "Separate items using NULL; recommended if your tool supports --read0 or similar").Default("false").Bool()
	errorOnNoSelection = picker.Flag("err-on-no-selection", "exit 1 when there is no selection").Default("false").Bool()

	clearer       = app.Command("clear", "Remove item/s from history")
	maxClearer    = clearer.Flag("max-items", "scrollview length").Default("15").Int()
	clearTool     = clearer.Flag("tool", "Which selector to use: wofi/bemenu/CUSTOM/dmenu/rofi/STDOUT").Short('t').String()
	clearToolArgs = clearer.Flag("tool-args", "Extra arguments to pass to the --tool").Short('T').Default("").String()
	clearAll      = clearer.Flag("all", "Remove all items").Short('a').Default("false").Bool()
	clearEsc      = clearer.Flag("print0", "Separate items using NULL; recommended if your tool supports --read0 or similar").Default("false").Bool()

	_ = app.Command("show-history", "Show all items from history")

	_ = app.Command("restore", "Serve the last recorded item from history")
)

func main() {
	app.Version(version)
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	action := kingpin.MustParse(app.Parse(os.Args[1:]))

	histfile, history, err := getHistory(*histpath)
	if err != nil {
		smartLog(err.Error(), "critical", *alert)
	}

	switch action {
	case "store":
		// read copy from stdin
		var stdin []string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(scanLines)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			smartLog("Couldn't get input from stdin.", "critical", *alert)
		}
		text := strings.Join(stdin, "")

		if *minChar > 0 && len(text) < *minChar {
			return
		}

		persist := !*noPersist
		if err := store(text, history, histfile, *maxDemon, persist); err != nil {
			smartLog(err.Error(), "critical", *alert)
		}
	case "pick":
		selection, err := selector(history, *maxPicker, *pickTool, "pick", *pickToolArgs, *pickEsc, *errorOnNoSelection)
		if err != nil {
			smartLog(err.Error(), "normal", *alert)
		}

		if selection != "" {
			// serve selection to the OS
			serveTxt(selection)
		}
	case "restore":
		if len(history) == 0 {
			fmt.Println("Nothing to restore")
			return
		}

		serveTxt(history[len(history)-1])
	case "show-history":
		if len(history) != 0 {
			urlsJSON, err := json.Marshal(history)
			if err != nil {
				fmt.Printf("Error marshalling history: %s\n", err.Error())
				return
			}

			fmt.Println(string(urlsJSON))
			return
		}
		fmt.Println("Nothing to show")
		return
	case "clear":
		// remove all history
		if *clearAll {
			if err := wipeAll(histfile); err != nil {
				smartLog(err.Error(), "normal", *alert)
			}
			return
		}

		if *clearTool == "" {
			fmt.Println("clipman: error: required flag --tool or --all not provided, try --help")
			os.Exit(1)
		}

		selection, err := selector(history, *maxClearer, *clearTool, "clear", *clearToolArgs, *clearEsc, *errorOnNoSelection)
		if err != nil {
			smartLog(err.Error(), "normal", *alert)
		}

		if selection == "" {
			return
		}

		if len(history) < 2 {
			// there was only one possible item we could select, and we selected it,
			// so wipe everything
			if err := wipeAll(histfile); err != nil {
				smartLog(err.Error(), "normal", *alert)
			}
			return
		}

		if selection == history[len(history)-1] {
			// wl-copy is still serving the copy, so replace with next latest
			// note: we alread exited if less than 2 items
			serveTxt(history[len(history)-2])
		}

		if err := write(filter(history, selection), histfile); err != nil {
			smartLog(err.Error(), "critical", *alert)
		}
	}
}

func wipeAll(histfile string) error {
	// clear WM's clipboard
	if err := exec.Command("wl-copy", "-c").Run(); err != nil {
		return err
	}

	if err := os.Remove(histfile); err != nil {
		return err
	}

	return nil
}

func getHistory(rawPath string) (string, []string, error) {
	// set histfile; expand user home
	histfile := rawPath
	if strings.HasPrefix(histfile, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil, err
		}
		histfile = strings.Replace(histfile, "~", home, 1)
	}

	// read history if it exists
	var history []string
	b, err := os.ReadFile(histfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("failure reading history file: %w", err)
		}
	} else {
		if err := json.Unmarshal(b, &history); err != nil {
			return "", nil, fmt.Errorf("failure parsing history: %w", err)
		}
	}

	return histfile, history, nil
}

func serveTxt(s string) {
	bin, err := exec.LookPath("wl-copy")
	if err != nil {
		smartLog(fmt.Sprintf("couldn't find wl-copy: %v\n", err), "low", *alert)
	}

	// daemonize wl-copy into a truly independent process
	// necessary for running stuff like `alacritty -e sh -c clipman pick`
	attr := &syscall.SysProcAttr{
		Setpgid: true,
	}

	// we mandate the mime type because we know we can only serve text; not doing this leads to weird bugs like #35
	if *primary {
		cmd := exec.Cmd{Path: bin, Args: []string{bin, "-p", "-t", "TEXT"}, Stdin: strings.NewReader(s), SysProcAttr: attr}
		if err := cmd.Run(); err != nil {
			smartLog(fmt.Sprintf("error running wl-copy -p: %s\n", err), "low", *alert)
		}
	} else {
		cmd := exec.Cmd{Path: bin, Args: []string{bin, "-t", "TEXT"}, Stdin: strings.NewReader(s), SysProcAttr: attr}
		if err := cmd.Run(); err != nil {
			smartLog(fmt.Sprintf("error running wl-copy: %s\n", err), "low", *alert)
		}
	}
}

// scanLines is a custom implementation of a split function for a bufio.Scanner.
// It has been modified from the standard library version to ensure that carriage return (\r)
// and newline (\n) characters are not dropped. This is important for maintaining the integrity
// of the input data, especially when dealing with text files or streams where these characters
// are significant.
//
// Parameters:
// - data: The byte slice to be scanned.
// - atEOF: A boolean indicating if the end of the file has been reached.
//
// Returns:
// - advance: The number of bytes to advance the input.
// - token: The next token to return to the user.
// - err: Any error encountered during scanning.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		b := data[0 : i+1]
		if *unix {
			b = dropCR(b)
		}
		return i + 1, b, nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		b := data
		if *unix {
			b = dropCR(b)
		}
		return len(data), b, nil
	}

	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data. This function has been modified from Go's
// standard library to ensure that carriage return (\r) characters are properly handled.
// It checks if the data ends with a newline (\n) and removes the preceding carriage return (\r)
// if present. This is useful for processing text data that may have different line ending
// conventions (e.g., Windows vs. Unix).
//
// Parameters:
// - data: The byte slice from which the terminal \r should be dropped.
//
// Returns:
// - A new byte slice with the terminal \r removed, if it was present.
func dropCR(data []byte) []byte {
	orig := data

	var lf bool
	if len(data) > 0 && data[len(data)-1] == '\n' {
		lf = true
		data = data[0 : len(data)-1]
	}

	if len(data) > 0 && data[len(data)-1] == '\r' {
		b := data[0 : len(data)-1]
		if lf {
			b = append(b, '\n')
		}
		return b
	}

	return orig
}
