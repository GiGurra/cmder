package util

import (
	"io"
)

type _tapper struct {
	w       io.Writer
	tapFunc func()
}

func (rw _tapper) Write(p []byte) (n int, err error) {
	rw.tapFunc()
	return rw.w.Write(p)
}

func newTapper(w io.Writer, tapFunc func()) io.Writer {
	return _tapper{
		w:       w,
		tapFunc: tapFunc,
	}
}

func TapWriterToChan(w io.Writer, tapChan chan any) io.Writer {
	return newTapper(w, func() { tapChan <- struct{}{} })
}
