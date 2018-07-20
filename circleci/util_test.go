package circleci

import (
	"testing"
)

func TestAccMaskCircleCiSecret(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "a",
			expected: "xxxx",
		},
		{
			input:    "aa",
			expected: "xxxxa",
		},
		{
			input:    "aaa",
			expected: "xxxxa",
		},
		{
			input:    "aaaa",
			expected: "xxxxaa",
		},
		{
			input:    "aaaaa",
			expected: "xxxxaa",
		},
		{
			input:    "aaaaaa",
			expected: "xxxxaaa",
		},
		{
			input:    "aaaaaaa",
			expected: "xxxxaaa",
		},
		{
			input:    "aaaaaaaa",
			expected: "xxxxaaaa",
		},
		{
			input:    "aaaaaaaaa",
			expected: "xxxxaaaa",
		},
		{
			input:    "aaaaaaaaaa",
			expected: "xxxxaaaa",
		},
		{
			input:    "aaaaaaaaaaa",
			expected: "xxxxaaaa",
		},
		{
			input:    "aaaaaaaaaaaa",
			expected: "xxxxaaaa",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {

			result := maskCircleCiSecret(tc.input)

			if result != tc.expected {
				t.Errorf("Mask was incorerect, got: %s, want: %s.", result, tc.expected)
			}

		})
	}
}
