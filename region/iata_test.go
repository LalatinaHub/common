package region_test

import (
	"testing"

	"github.com/LalatinaHub/common/region"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		code     string
		expected string
		found    bool
	}{
		{"CGK", "JAKARTA-CENGKARENG", true},
		{"cgk", "JAKARTA-CENGKARENG", true},
		{" SIN ", "SINGAPORE", true},
		{"NRT", "TOKYO", true},
		{"JFK", "NEW YORK", true},
		{"NONEXISTENT", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result, ok := region.Lookup(tt.code)
			assert.Equal(t, tt.found, ok)
			assert.Equal(t, tt.expected, result)
		})
	}
}
