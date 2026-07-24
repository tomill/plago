package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	l := List{}
	err := l.UnmarshalText([]byte(`
ch1 # channel name
ch2
ch3, ch4`,
	))
	assert.NoError(t, err)
	assert.Equal(t, List{"ch1", "ch2", "ch3", "ch4"}, l)
}
