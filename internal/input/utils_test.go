package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cached := &cache[int]{}

	dummy := 0
	callback := func() int {
		dummy += 1
		return dummy
	}

	assert.Equal(t, 1, cached.get("one", callback))
	assert.Equal(t, 1, cached.get("one", callback), "memo")
	assert.Equal(t, 2, cached.get("two", callback))
}
