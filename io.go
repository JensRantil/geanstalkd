package main

import (
	"bufio"
	"errors"
)

var errLineTooLong = errors.New("Line was too long.")

func readCappedLine(r *bufio.Reader, maxBytes int) (string, error) {
	var line []byte
	for {
		l, more, err := r.ReadLine()
		if err != nil {
			return "", err
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return string(l), nil
		}
		line = append(line, l...)
		if len(line) > maxBytes {
			return string(line), errLineTooLong
		}
		if !more {
			break
		}
	}
	return string(line), nil
}
