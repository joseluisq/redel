package redel

import (
	"bufio"
	"bytes"
	"io"
)

// Redel provides an interface (around Scanner) for replace string occurrences
// between two string delimiters.
type Redel struct {
	r           io.Reader // The reader provided by the client.
	start       string    // Start string delimiter.
	end         string    // End string delimiter.
	replacement string    // Replacement string.
}

// EOF is a byte used for determine the EOF in scanning.
var EOF = []byte("eof")

// ScannerFilterFunc is the callback filter function that will be called per replacement match
// which allows to control the processing's replacement (true or false)
type ScannerFilterFunc func(matchValue []byte, atEOF bool) bool

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

// Replace function scanns and replaces string occurrences
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

// FilterReplace function scanns and replaces string occurrences but using a filter function
// to control the processing of every replacement (true or false)
func (rd *Redel) FilterReplace(filterFunc ScannerFilterFunc, scannerFn ScannerFunc) {
	scanner := bufio.NewScanner(rd.r)

	ScanByDelimiters := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		endLen := len(rd.end)
		startLen := len(rd.start)

		if (startLen <= 0 || endLen <= 0) || (atEOF && len(data) == 0) {
			return 0, nil, nil
		}

		if from := bytes.Index(data, []byte(rd.start)); from >= 0 {
			if to := bytes.Index(data[from:], []byte(rd.end)); to >= 0 {
				a := from + startLen
				b := from + endLen + (to - endLen)
				val := data[a:b]

				if filterFunc(val, atEOF) {
					return b, data[0:a], nil
				}
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
			scannerFn(wchunk, true)
		} else {
			wchunk = append(chunk, rd.replacement...)
			scannerFn(wchunk, false)
		}
	}
}
