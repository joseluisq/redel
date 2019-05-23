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

// redelValues interface contains replacement values
type redelValues struct {
	start int    // Start index.
	end   int    // End index.
	value []byte // Value to replace.
}

// redelValues interface contains replacement values
type redelValidator struct {
	from      int
	to        int
	validFrom bool
	validTo   bool
}

// EOF is a byte used for determine the EOF in scanning.
var EOF = []byte("eof")

// FilterReplacementWithFunc is the callback filter function that will be called
// per replacement match which allows to control the processing and return a custom replace
type FilterReplacementWithFunc func(matchValue []byte) []byte

// FilterReplacementFunc is the callback filter function that will be called per replacement match
// which allows to control the processing's replacement (true or false)
type FilterReplacementFunc func(matchValue []byte) bool

// ReplacementMapFunc is the callback function that will be called
// for every successful replacement.
type ReplacementMapFunc func(data []byte, atEOF bool)

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

// Replace function scans and replaces string occurrences
// for the privided delimiters.
// Replace requires a callback function that will be called
// for every successful replacement.
// The callback will receive two params:
// - data []byte (Each successful replaced byte)
// - atEOF bool (If loop is EOF)
func (rd *Redel) Replace(replaceFunc ReplacementMapFunc) {
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
			replaceFunc(wchunk, true)
		} else {
			wchunk = append(chunk, rd.replacement...)
			replaceFunc(wchunk, false)
		}
	}
}

// FilterReplaceFull function scans and replaces string occurrences but using a filter function
// to control the processing of every replacement
func (rd *Redel) FilterReplaceFull(replaceMapFunc ReplacementMapFunc, filterFunc FilterReplacementWithFunc, preserveDelimiters bool, replaceWith bool) {
	scanner := bufio.NewScanner(rd.r)
	var valuesData []redelValues

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

				if !preserveDelimiters {
					a = from
					b = from + endLen + to
				}

				v := data[a:b]

				if len(v) > 0 {
					valuesData = append(valuesData, redelValues{
						start: a,
						end:   b,
						value: v,
					})
				}

				return b, data[0:a], nil
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
		bytesR := scanner.Bytes()
		atEOF := bytes.HasSuffix(bytesR, EOF)

		valuesLen := len(valuesData) - 1

		var value []byte
		var valueToReplace []byte

		if !atEOF && valuesLen >= 0 {
			value = valuesData[valuesLen].value
			valueToReplace = filterFunc(value)
		}

		bytesW := bytesR

		if atEOF {
			bytesW = bytes.Split(bytesR, EOF)[0]
		} else {
			if replaceWith {
				bytesW = append(bytesR, valueToReplace...)
			} else {
				if len(valueToReplace) > 0 && string(valueToReplace) == "0" {
					bytesW = append(bytesR, value...)
				} else {
					bytesW = append(bytesR, rd.replacement...)
				}
			}
		}

		replaceMapFunc(bytesW, atEOF)
	}
}

// FilterReplace function scans and replaces string occurrences controling the processing of every replacement via boolean return value
func (rd *Redel) FilterReplace(replaceMapFunc ReplacementMapFunc, filterFunc FilterReplacementFunc, preserveDelimiters bool) {
	rd.FilterReplaceFull(replaceMapFunc, func(matchValue []byte) []byte {
		result := []byte("0")

		ok := filterFunc(matchValue)

		if ok {
			result = []byte("1")
		}

		return result
	}, preserveDelimiters, false)
}

// FilterReplaceWith function scans and replaces string occurrences controling the processing of every replacement via a custom returned []byte replacement
func (rd *Redel) FilterReplaceWith(replaceMapFunc ReplacementMapFunc, filterReplaceWithFunc FilterReplacementWithFunc, preserveDelimiters bool) {
	rd.FilterReplaceFull(replaceMapFunc, filterReplaceWithFunc, preserveDelimiters, true)
}
