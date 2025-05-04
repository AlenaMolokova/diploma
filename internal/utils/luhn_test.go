package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhnCheck(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "valid number",
			number:   "4532015112830366",
			expected: true,
		},
		{
			name:     "invalid number",
			number:   "4532015112830367",
			expected: false,
		},
		{
			name:     "empty string",
			number:   "",
			expected: false,
		},
		{
			name:     "non-digit characters",
			number:   "4532a15112830366",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LuhnCheck(tt.number)
			assert.Equal(t, tt.expected, result, "LuhnCheck result mismatch")
		})
	}
}
