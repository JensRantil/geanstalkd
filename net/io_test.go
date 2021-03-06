package net

import (
	"bufio"
	"bytes"
	. "testing"
)

func TestReadingTooLongCappedLine(t *T) {
	t.Parallel()

	var b bytes.Buffer
	for i := 0; i < 10000; i++ {
		b.WriteByte(42)
	}
	reader := bufio.NewReader(&b)
	_, err := readCappedLine(reader, 1024)

	if err != errLineTooLong {
		t.Error("Expected the function to return errLineTooLong.")
	}
}

func TestReadingShortCappedLine(t *T) {
	t.Parallel()

	var b bytes.Buffer
	b.WriteString("Hello!\n")
	reader := bufio.NewReader(&b)
	_, err := readCappedLine(reader, 1024)

	if err != nil {
		t.Error("Did not expect an error.")
	}
}
