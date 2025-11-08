package trace

import (
	"bytes"
	"io"
	"testing"
)

func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)

	if tracer == nil {
		t.Error("Return from New should not be nil!")
	} else {
		tracer.Trace("Hello trace package.")

		if buf.String() != "Hello trace package.\n" {
			t.Errorf("Trace should not write '%s'.", buf.String())
		}
	}

}
