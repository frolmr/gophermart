package luhn

import (
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid Luhn number",
			number:   "79927398713",
			expected: true,
		},
		{
			name:     "Invalid Luhn number",
			number:   "79927398712",
			expected: false,
		},
		{
			name:     "Valid number with non-digit characters",
			number:   "7992-739-8713",
			expected: false,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: false,
		},
		{
			name:     "Single digit",
			number:   "0",
			expected: true,
		},
		{
			name:     "Valid number with spaces",
			number:   "7992 7398 713",
			expected: true,
		},
		{
			name:     "Number with leading zeros",
			number:   "00411886993124",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Check(tt.number)
			if result != tt.expected {
				t.Errorf("Check(%q) = %v; expected %v", tt.number, result, tt.expected)
			}
		})
	}
}
