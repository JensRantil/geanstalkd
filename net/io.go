package net

import (
	"bufio"
	"bytes"
	"errors"
)

var errLineTooLong = errors.New("line too long")

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
