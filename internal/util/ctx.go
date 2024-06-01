package util

import (
	"io"
)

type _observer struct {
	fn func()
}

func (rw _observer) Write(p []byte) (n int, err error) {
	rw.fn()
	return len(p), nil
}

func newObserver(fn func()) io.Writer {
	return _observer{
		fn: fn,
	}
}

func NewObserver(ch chan any) io.Writer {
	return newObserver(func() { ch <- struct{}{} })
}
