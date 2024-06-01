package util

import (
	"io"
)

type ResetFunc func()

type ResetWriter interface {
	io.Writer
}

type resetWriter struct {
	w         io.Writer
	resetFunc ResetFunc
}

func (rw resetWriter) Write(p []byte) (n int, err error) {
	rw.resetFunc()
	return rw.w.Write(p)
}

func NewResetWriter(w io.Writer, resetFunc ResetFunc) ResetWriter {
	return resetWriter{
		w:         w,
		resetFunc: resetFunc,
	}
}

func NewResetWriterCh(w io.Writer, resetChan chan any) ResetWriter {
	return NewResetWriter(w, func() { resetChan <- struct{}{} })
}
