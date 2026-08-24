package config

import (
	"strings"
	"unicode"
)

type List []string

func (l *List) UnmarshalText(b []byte) error {
	for v := range strings.SplitSeq(string(b), "\n") {
		v, _, _ = strings.Cut(v, "#")

		*l = append(*l, strings.FieldsFunc(v, func(r rune) bool {
			return unicode.IsSpace(r) || r == ','
		})...)
	}

	return nil
}
