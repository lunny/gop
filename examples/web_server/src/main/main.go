package main

import (
	"io"

	"github.com/lunny/tango"
)

func main() {
	t := tango.Classic()
	t.Get("/", func(t *tango.Context) {
		io.WriteString(t, "hello tango")
	})
}
