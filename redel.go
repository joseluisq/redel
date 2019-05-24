package redel

import (
	"bufio"
	"bytes"
	"io"
)

type (
	// Redel provides an interface (around Scanner) for replace string occurrences
	// between two string delimiters.
	Redel struct {
		ioReader       io.Reader // The reader provided by the client.
		startDelimiter string    // Start string delimiter.
		endDelimiter   string    // End string delimiter.
	}

	// redelValues interface contains replacement values
	redelValues struct {
		startDelimiter int    // Start index.
		endDelimiter   int    // End index.
		value          []byte // Value to replace.
	}

	// FilterValueFunc defines a filter function that will be called per replacement
	// which supports a return `bool` value to apply the replacement or not.
	FilterValueFunc func(matchValue []byte) bool

	// FilterValueReplaceFunc defines a filter function that will be called per replacement
	// which supports a return `[]byte` value to customize the replacement.
	FilterValueReplaceFunc func(matchValue []byte) []byte

	// ReplacementMapFunc defines a map function that will be called per scan token.
	ReplacementMapFunc func(data []byte, atEOF bool)
)

var (
	// EOF is an byte intended to determine the EOF in scanning.
	EOF = []byte("eof")
)

// New creates a new Redel instance.
// - ioReader io.Reader (Input reader)
// - startDelimiter string (Start string delimiter)
// - endDelimiter string (End string delimiter)
func New(ioReader io.Reader, startDelimiter string, endDelimiter string) *Redel {
	return &Redel{
		ioReader:       ioReader,
		startDelimiter: startDelimiter,
		endDelimiter:   endDelimiter,
	}
}

// replaceFilterFunc API function which scans and replace string by supporting different options.
// It's used by API's replace methods.
func (rd *Redel) replaceFilterFunc(
	replacementMapFunc ReplacementMapFunc,
	filterFunc FilterValueReplaceFunc,
	preserveDelimiters bool,
	replaceWith bool,
	replacement []byte,
) {
	scanner := bufio.NewScanner(rd.ioReader)
	var valuesData []redelValues

	ScanByDelimiters := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		endLen := len(rd.endDelimiter)
		startLen := len(rd.startDelimiter)

		if (startLen <= 0 || endLen <= 0) || (atEOF && len(data) == 0) {
			return 0, nil, nil
		}

		if from := bytes.Index(data, []byte(rd.startDelimiter)); from >= 0 {
			if to := bytes.Index(data[from:], []byte(rd.endDelimiter)); to >= 0 {
				a := from + startLen
				b := from + endLen + (to - endLen)

				v := data[a:b]

				if len(v) > 0 {
					valuesData = append(valuesData, redelValues{
						startDelimiter: a,
						endDelimiter:   b,
						value:          v,
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

	counterDelimiterStart := false

	for scanner.Scan() {
		bytesR := append([]byte(nil), scanner.Bytes()...)
		atEOF := bytes.HasSuffix(bytesR, EOF)

		valuesLen := len(valuesData) - 1
		value := []byte(nil)
		valueToReplace := []byte(nil)

		if !atEOF && valuesLen >= 0 {
			value = append([]byte(nil), valuesData[valuesLen].value...)
			valueToReplace = filterFunc(value)
		}

		bytesW := bytesR

		if atEOF {
			bytesW = bytes.Split(bytesR, EOF)[0]
		} else {
			if replaceWith {
				// takes the callback value instead
				bytesW = append(bytesR, valueToReplace...)
			} else {
				// don't replace and use the value instead
				if len(valueToReplace) == 0 {
					// takes the array value instead
					bytesW = append(bytesR, value...)
				} else {
					bytesW = append(bytesR, replacement...)
				}
			}
		}

		if !preserveDelimiters {
			if counterDelimiterStart && bytes.Index(bytesW, []byte(rd.endDelimiter)) >= 0 {
				bytesW = bytes.Replace(bytesW, []byte(rd.endDelimiter), []byte(nil), 1)
				counterDelimiterStart = false
			}

			if !counterDelimiterStart && bytes.Index(bytesW, []byte(rd.startDelimiter)) >= 0 {
				bytesW = bytes.Replace(bytesW, []byte(rd.startDelimiter), []byte(nil), 1)
				counterDelimiterStart = true
			}
		}

		replacementMapFunc(bytesW, atEOF)
	}
}

// Replace function replaces every occurrence with a custom replacement token.
func (rd *Redel) Replace(replacement string, replacementMapFunc ReplacementMapFunc) {
	rd.ReplaceFilterWith(replacementMapFunc, func(value []byte) []byte {
		return []byte(replacement)
	}, false)
}

// ReplaceFilter function scans and replaces string occurrences
// filtering replacement values via a return `bool` value.
func (rd *Redel) ReplaceFilter(
	replacement []byte,
	replacementMapFunc ReplacementMapFunc,
	filterFunc FilterValueFunc,
	preserveDelimiters bool,
) {
	rd.replaceFilterFunc(replacementMapFunc, func(matchValue []byte) []byte {
		result := []byte(nil)

		ok := filterFunc(matchValue)

		if ok {
			result = []byte("1")
		}

		return result
	}, preserveDelimiters, false, replacement)
}

// ReplaceFilterWith function scans and replaces string occurrences via a custom replacement callback.
func (rd *Redel) ReplaceFilterWith(
	mapFunc ReplacementMapFunc,
	filterReplaceFunc FilterValueReplaceFunc,
	preserveDelimiters bool,
) {
	rd.replaceFilterFunc(mapFunc, filterReplaceFunc, preserveDelimiters, true, []byte(nil))
}
