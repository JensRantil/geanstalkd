package net

import (
	"bufio"
	"bytes"
	"errors"
)

var errLineTooLong = errors.New("line too long")

// readCappedLine reads a line from a bufio.Reader, but sets an upper limit to
// how much can be read to avoid DDoS attacks when reading from network.
func readCappedLine(r *bufio.Reader, maxBytes int) (string, error) {
	line := bytes.Buffer{}

	var l []byte
	var more bool
	var err error
	for {
		l, more, err = r.ReadLine()
		if err != nil {
			break
		}
		line.Write(l)
		if line.Len() > maxBytes {
			err = errLineTooLong
			break
		}
		if !more {
			break
		}
	}
	return line.String(), nil
}
