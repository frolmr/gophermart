package formatter

import (
	"testing"
)

func TestConvertToCurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected float64
	}{
		{"Zero value", 0, 0.0},
		{"Positive value", 100, 1.0},
		{"Large value", 123456789, 1234567.89},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToCurrency(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToCurrency(%d) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToSubunit(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected int
	}{
		{"Zero value", 0.0, 0},
		{"Positive value", 1.0, 100},
		{"Fractional value", 123.456, 12346},
		{"Large value", 1234567.89, 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToSubunit(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToSubunit(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
