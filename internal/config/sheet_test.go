package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSheet(t *testing.T) {
	if os.Getenv("SHEET_CREDENTIALS") == "" {
		t.Skip("env SHEET_CREDENTIALS not set")
	}

	type Detective struct {
		Author  string
		Name    string
		Created int
	}

	var actual Sheet[Detective]
	err := actual.UnmarshalText([]byte("https://docs.google.com/spreadsheets/d/1WUpe_O9Vs623EPuELGudnNBZgE-C1rTtY_jPJ_9XJyQ/edit?gid=267698540#gid=267698540"))
	assert.NoError(t, err)

	expected := []Detective{
		{Author: "横溝正史", Name: "金田一耕助", Created: 0},
		{Author: "青山剛昌", Name: "江戸川コナン", Created: 0},
	}

	assert.Equal(t, expected, actual)
}
