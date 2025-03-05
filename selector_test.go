package main

import (
	"testing"

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
				"北京",                                   // Chinese
				"こんにちは",                                // Japanese
				"안녕하세요",                                // Korean
				"Москва",                               // Cyrillic
				"ÄÖÜß",                                 // German
				"Ελληνικά",                             // Greek
				"\u0915\u093e\u0928\u092a\u0941\u0930", // Devanagari
			},
			// These should remain identical after NFC normalization
			expected: []string{
				"北京",
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
