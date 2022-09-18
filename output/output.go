package output

import (
	"fmt"

	"github.com/Code-Hex/dd"
	"github.com/tomill/centre/message"
)

type Flusher interface {
	Flush(*message.Timeline) error
}

type Stdout struct{}

func (p Stdout) Flush(timeline *message.Timeline) error {
	fmt.Println(timeline.PlainText())
	return nil
}

type Dump struct{}

func (p Dump) Flush(timeline *message.Timeline) error {
	fmt.Println(dd.Dump(timeline))
	return nil
}
