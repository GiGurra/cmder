package util

import (
	"bytes"
	"io"
	"testing"
)

func TestTapWriterToChan(t *testing.T) {
	sink := bytes.Buffer{}
	tap := make(chan any, 10)
	obs := NewSignalForwarderWriter(tap)
	combined := io.MultiWriter(&sink, obs)

	_, _ = combined.Write([]byte("hello"))

	if sink.String() != "hello" {
		t.Error("Expected 'hello', got", sink.String())
	}

	// Test that the tap channel is written to
	select {
	case x := <-tap:
		if x != struct{}{} {
			t.Error("Expected tap channel to be written to")
		}
	default:
		t.Error("Expected tap channel to be written to")
	}
}
