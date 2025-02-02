package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterByCSV(t *testing.T) {
	s := NewSlices()

	tests := []struct {
		name   string
		slice  []string
		csv    string
		expect []string
	}{
		{
			name:   "No CSV filter",
			slice:  []string{"apple", "banana", "cherry"},
			csv:    "",
			expect: []string{"apple", "banana", "cherry"},
		},
		{
			name:   "Filter single item",
			slice:  []string{"apple", "banana", "cherry"},
			csv:    "banana",
			expect: []string{"banana"},
		},
		{
			name:   "Filter multiple items",
			slice:  []string{"apple", "banana", "cherry"},
			csv:    "banana, cherry",
			expect: []string{"banana", "cherry"},
		},
		{
			name:   "Filter with spaces",
			slice:  []string{"apple", "banana", "cherry"},
			csv:    " banana , cherry ",
			expect: []string{"banana", "cherry"},
		},
		{
			name:   "No match",
			slice:  []string{"apple", "banana", "cherry"},
			csv:    "grape",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.FilterByCSV(tt.slice, tt.csv)
			assert.Equal(t, tt.expect, result)
		})
	}
}
