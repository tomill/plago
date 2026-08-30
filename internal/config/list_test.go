package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{
			input: `
ch1 # channel name
	ch2
ch3, ch4`,
			expected: []string{"ch1", "ch2", "ch3", "ch4"}},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			l := List{}
			err := l.UnmarshalText([]byte(test.input))
			assert.NoError(t, err)
			assert.Equal(t, List(test.expected), l)
		})
	}
}
