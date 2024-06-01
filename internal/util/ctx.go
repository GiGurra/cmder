package util

import (
	"io"
)

type _sfw struct {
	fn func()
}

func (rw _sfw) Write(p []byte) (n int, err error) {
	rw.fn()
	return len(p), nil
}

func newSFW(fn func()) io.Writer {
	return _sfw{
		fn: fn,
	}
}

func NewSignalForwarderWriter(ch chan any) io.Writer {
	return newSFW(func() { ch <- struct{}{} })
}
