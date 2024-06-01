package util

import (
	"bytes"
	"testing"
)

func TestTapWriterToChan(t *testing.T) {
	sink := bytes.Buffer{}
	tap := make(chan any, 10)
	tapper := TapWriterToChan(&sink, tap)

	_, _ = tapper.Write([]byte("hello"))

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
