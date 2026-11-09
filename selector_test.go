package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func TestPreprocessDataNormalization(t *testing.T) {
	// Test cases with different Unicode compositions
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "NFC vs NFD",
			// "é" can be represented as single code point (NFC) or as "e" + combining accent (NFD)
			input: []string{
				"café",               // NFC form
				"cafe\u0301",         // NFD form (e + combining accent)
				"résumé",             // NFC form
				"re\u0301sume\u0301", // NFD form
			},
			expected: []string{
				"café", // All should be normalized to NFC
				"café",
				"résumé",
				"résumé",
			},
		},
		{
			name: "Special Characters",
			input: []string{
				"北京",                                   //nolint:gosmopolitan // Chinese
				"こんにちは",                                // Japanese
				"안녕하세요",                                // Korean
				"Москва",                               // Cyrillic
				"ÄÖÜß",                                 // German
				"Ελληνικά",                             // Greek
				"\u0915\u093e\u0928\u092a\u0941\u0930", // Devanagari
			},
			// These should remain identical after NFC normalization
			expected: []string{
				"北京", //nolint:gosmopolitan
				"こんにちは",
				"안녕하세요",
				"Москва",
				"ÄÖÜß",
				"Ελληνικά",
				"\u0915\u093e\u0928\u092a\u0941\u0930",
			},
		},
		{
			name: "Edge Cases",
			input: []string{
				"\u1e9b\u0323",       // NFC: ẛ̣ (Latin small letter long s with dot below)
				"\u0073\u0323\u0307", // NFD: ṩ (s + dot below + dot above)
			},
			expected: []string{
				"\u1e9b\u0323",
				"\u1e69", // Should normalize to single code point
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run with normalization enabled
			processed, _ := preprocessData(tc.input, 0, false, true)

			// Reverse the processed data to match the original order (since preprocessData reverses)
			reversed := make([]string, len(processed))
			for i := 0; i < len(processed); i++ {
				reversed[i] = processed[len(processed)-1-i]
			}

			// Verify each string is in NFC form
			for i, str := range reversed {
				// Check if the result is in NFC form
				if !norm.NFC.IsNormalString(str) {
					t.Errorf("Result not in NFC form: %q", str)
				}

				// Check if the normalized form matches what we expect
				if str != tc.expected[i] {
					t.Errorf("Expected %q, got %q", tc.expected[i], str)
				}
			}
		})
	}
}

func TestPreprocessDataTruncationUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxChars int
		want     string
	}{
		{
			name:     "ASCII fits exactly",
			input:    "abc",
			maxChars: 3,
			want:     "abc",
		},
		{
			name:     "ASCII truncated at boundary",
			input:    "abcdef",
			maxChars: 3,
			want:     "abc",
		},
		{
			name:     "2-byte rune fits exactly",
			input:    "aébc", // 61 C3 A9 62 63
			maxChars: 3,
			want:     "aé",
		},
		{
			name:     "2-byte rune straddles limit",
			input:    "aébc",
			maxChars: 2,
			want:     "a", // must back off from the split é
		},
		{
			name:     "4-byte rune straddles limit",
			input:    "a😀b", // 61 F0 9F 98 80 62
			maxChars: 2,
			want:     "a",
		},
		{
			name:     "4-byte rune fits exactly",
			input:    "a😀b",
			maxChars: 5,
			want:     "a😀",
		},
		{
			name:     "3-byte runes truncated at boundary",
			input:    "\u65e5\u672c\u8a9e",
			maxChars: 3,
			want:     "\u65e5",
		},
		{
			name:     "3-byte runes cut mid-rune",
			input:    "\u65e5\u672c\u8a9e",
			maxChars: 5,
			want:     "\u65e5", // 5 bytes would split the second rune
		},
		{
			name:     "first rune larger than limit",
			input:    "é",
			maxChars: 1,
			want:     "", // cannot include any complete rune
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, guide := preprocessData([]string{tt.input}, tt.maxChars, false, false)
			if len(got) != 1 {
				t.Fatalf("preprocessData returned %d results, want 1", len(got))
			}
			if got[0] != tt.want {
				t.Errorf("truncated = %q, want %q", got[0], tt.want)
			}
			if !utf8.ValidString(got[0]) {
				t.Errorf("truncated result is not valid UTF-8: %q", got[0])
			}
			if len(got[0]) > tt.maxChars {
				t.Errorf("truncated result exceeds maxChars=%d: %q (len %d)",
					tt.maxChars, got[0], len(got[0]))
			}
			if !strings.HasPrefix(tt.input, got[0]) {
				t.Errorf("truncated result %q is not a prefix of input %q", got[0], tt.input)
			}
			if orig, ok := guide[got[0]]; !ok || orig != tt.input {
				t.Errorf("guide[%q] = %q (ok=%v), want %q", got[0], orig, ok, tt.input)
			}
		})
	}
}

func TestPreprocessDataTruncationAlwaysValidUTF8(t *testing.T) {
	inputs := []string{
		"aébc",
		"a😀b",
		"\u65e5\u672c\u8a9e\u30c6\u30b9\u30c8",
		"e\u0301", // NFD é (e + combining acute)
		"𝕳𝖊𝖑𝖑𝖔",   // astral plane math letters
		"ÄÖÜäöüß", // 2-byte Latin-1 supplement
		"",        // empty input
	}

	for _, input := range inputs {
		for maxChars := 1; maxChars <= len(input)+2; maxChars++ {
			got, _ := preprocessData([]string{input}, maxChars, false, false)
			if len(got) != 1 {
				t.Fatalf("preprocessData(%q, %d) returned %d results", input, maxChars, len(got))
			}
			result := got[0]

			if !utf8.ValidString(result) {
				t.Errorf("invalid UTF-8 for input=%q maxChars=%d: got %q", input, maxChars, result)
			}
			if len(result) > maxChars {
				t.Errorf("exceeds maxChars for input=%q maxChars=%d: got %q (len %d)",
					input, maxChars, result, len(result))
			}
			if !strings.HasPrefix(input, result) {
				t.Errorf("not a prefix for input=%q maxChars=%d: got %q", input, maxChars, result)
			}
		}
	}
}
