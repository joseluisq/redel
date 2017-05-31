package redel

import (
	"bufio"
	"bytes"
	"io"
)

// Redel provides an interface (around Scanner) for replace strings occurrences
// between two strings delimiters.
type Redel struct {
	r           io.Reader // The reader provided by the client.
	start       string    // Start string delimiter.
	end         string    // End string delimiter.
	replacement string    // Replacement string.
}

// EOF is a byte used for determine the EOF in scanning.
var EOF = []byte("eof")

// ScannerFunc is the callback function that will be called
// for every successful replacement.
type ScannerFunc func(data []byte, atEOF bool)

// NewRedel returns a new Redel to read from r.
// - r io.Reader (Input reader)
// - start string (Start string delimiter)
// - end string (End string delimiter)
// - replacement string (Replacement string)
func NewRedel(r io.Reader, start string, end string, replacement string) *Redel {
	return &Redel{
		r:           r,
		start:       start,
		end:         end,
		replacement: replacement,
	}
}

// Replace function scanns and replaces strings occurrences
// for the privided delimiters.
// Replace requires a callback function that will be called
// for every successful replacement.
// The callback will receive two params:
// - data []byte (Each successful replaced byte)
// - atEOF bool (If loop is EOF)
func (rd *Redel) Replace(callback ScannerFunc) {
	scanner := bufio.NewScanner(rd.r)

	ScanByDelimiters := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, []byte(rd.start)); i >= 0 {
			if len(rd.end) > 1 {
				if b := bytes.Index(data, []byte(rd.end)); b >= 0 {
					return b + len(rd.end), data[0:i], nil
				}
			} else if len(rd.start) > 1 {
				return i + len(rd.start), data[0:i], nil
			}
		}

		if atEOF && len(data) > 0 {
			last := append(data[0:], EOF...)
			return len(data), last, nil
		}

		return 0, nil, nil
	}

	scanner.Split(ScanByDelimiters)

	for scanner.Scan() {
		var wchunk []byte
		var chunk = scanner.Bytes()

		if bytes.HasSuffix(chunk, EOF) {
			wchunk = bytes.Split(chunk, EOF)[0]
			callback(wchunk, true)
		} else {
			wchunk = append(chunk, rd.replacement...)
			callback(wchunk, false)
		}
	}
}
