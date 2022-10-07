package output

import (
	"io"
)

type ConditionalWriter struct {
	Writer io.Writer
	ok     bool
}

// Write writes b only if MuteableWriter.Mute is false,
// otherwise it is no-op.
func (w ConditionalWriter) Write(b []byte) (int, error) {
	if !w.ok {
		return 0, nil
	}
	return w.Writer.Write(b)
}

func (w ConditionalWriter) If(ok bool) ConditionalWriter {
	return ConditionalWriter{
		Writer: w.Writer,
		ok:     ok,
	}
}

func (w ConditionalWriter) Or(ok bool) ConditionalWriter {
	return ConditionalWriter{
		Writer: w.Writer,
		ok:     w.ok || ok,
	}
}
