// storer_test.go
package main

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		text     string
		maxItems int
		want     []string
	}{
		{
			name:     "empty history, new item",
			initial:  nil,
			text:     "A",
			maxItems: 3,
			want:     []string{"A"},
		},
		{
			name:     "duplicate of last item is ignored",
			initial:  []string{"A", "B"},
			text:     "B",
			maxItems: 3,
			want:     []string{"A", "B"},
		},
		{
			name:     "re-copy non-last item must not drop oldest",
			initial:  []string{"A", "B", "C"},
			text:     "B",
			maxItems: 3,
			want:     []string{"A", "C", "B"},
		},
		{
			name:     "new item at capacity evicts oldest",
			initial:  []string{"A", "B", "C"},
			text:     "D",
			maxItems: 3,
			want:     []string{"B", "C", "D"},
		},
		{
			name:     "duplicate not-last item in non-full history",
			initial:  []string{"A", "B"},
			text:     "A",
			maxItems: 3,
			want:     []string{"B", "A"},
		},
		{
			name:     "max-items=0 means unlimited",
			initial:  []string{"A", "B", "C"},
			text:     "D",
			maxItems: 0,
			want:     []string{"A", "B", "C", "D"},
		},
		{
			name:     "empty text is a no-op",
			initial:  []string{"A"},
			text:     "",
			maxItems: 3,
			want:     []string{"A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			histfile := filepath.Join(t.TempDir(), "history.json")

			// restore initial state
			if tt.initial != nil {
				if err := write(tt.initial, histfile); err != nil {
					t.Fatalf("setup write: %v", err)
				}
			}

			_, history, err := getHistory(histfile)
			if err != nil {
				t.Fatalf("getHistory: %v", err)
			}

			if err := store(tt.text, history, histfile, tt.maxItems, false); err != nil {
				t.Fatalf("store: %v", err)
			}

			_, got, err := getHistory(histfile)
			if err != nil {
				t.Fatalf("getHistory after: %v", err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("history = %v, want %v", got, tt.want)
			}
		})
	}
}
