package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func store(text string, history []string, histfile string, maxChar int, persist bool) error {
	if text == "" {
		return nil
	}

	// skip if the text is already the latest entry; this avoids an endless
	// loop, see https://github.com/bugaevc/wl-clipboard/issues/65
	if l := len(history); l > 0 && history[l-1] == text {
		return nil
	}

	// remove duplicates
	history = filter(history, text)

	// drop oldest items that exceed max list size
	// if max = 0, we allow infinite history; NOTE: users should NOT rely on this behaviour as we might change it without notice
	if maxChar != 0 && len(history) >= maxChar {
		history = history[len(history)-maxChar+1:]
	}

	history = append(history, text)

	// dump history to file so that other apps can query it
	if err := write(history, histfile); err != nil {
		return fmt.Errorf("error writing history: %w", err)
	}

	// make the copy buffer available to all applications,
	// even when the source has disappeared
	if persist {
		serveTxt(text)
	}

	return nil
}

// filter removes all occurrences of text.
func filter(slice []string, text string) []string {
	var filtered []string
	for _, s := range slice {
		if s != text {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

// write dumps history to json file.
func write(history []string, histfile string) error {
	b, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return os.WriteFile(histfile, b, 0o600)
}
