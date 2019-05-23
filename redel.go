package redel

import (
	"bufio"
	"bytes"
	"io"
)

// Redel provides an interface (around Scanner) for replace string occurrences
// between two string delimiters.
type Redel struct {
	r     io.Reader // The reader provided by the client.
	start string    // Start string delimiter.
	end   string    // End string delimiter.
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
func NewRedel(r io.Reader, start string, end string) *Redel {
	return &Redel{
		r:     r,
		start: start,
		end:   end,
	}
}

// Replace function replaces every occurrence with a custom replacement token
func (rd *Redel) Replace(replacement []byte, replaceMapFunc ReplacementMapFunc) {
	rd.FilterReplaceWith(replaceMapFunc, func(value []byte) []byte {
		return replacement
	}, false)
}

// FilterReplaceFull function scans and replaces string occurrences but using a filter function
// to control the processing of every replacement
func (rd *Redel) FilterReplaceFull(
	replaceMapFunc ReplacementMapFunc,
	filterFunc FilterReplacementWithFunc,
	preserveDelimiters bool,
	replaceWith bool,
	replacement []byte,
) {
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
			if counterDelimiterStart && bytes.Index(bytesW, []byte(rd.end)) >= 0 {
				bytesW = bytes.Replace(bytesW, []byte(rd.end), []byte(nil), 1)
				counterDelimiterStart = false
			}

			if !counterDelimiterStart && bytes.Index(bytesW, []byte(rd.start)) >= 0 {
				bytesW = bytes.Replace(bytesW, []byte(rd.start), []byte(nil), 1)
				counterDelimiterStart = true
			}
		}

		replaceMapFunc(bytesW, atEOF)
	}
}

// FilterReplace function scans and replaces string occurrences via a `bool` callback
func (rd *Redel) FilterReplace(
	replacement []byte,
	replaceMapFunc ReplacementMapFunc,
	filterFunc FilterReplacementFunc,
	preserveDelimiters bool,
) {
	rd.FilterReplaceFull(replaceMapFunc, func(matchValue []byte) []byte {
		result := []byte(nil)

		ok := filterFunc(matchValue)

		if ok {
			result = []byte("1")
		}

		return result
	}, preserveDelimiters, false, replacement)
}

// FilterReplaceWith function scans and replaces string occurrences via a custom `[]byte` callback
func (rd *Redel) FilterReplaceWith(
	replaceMapFunc ReplacementMapFunc,
	filterReplaceWithFunc FilterReplacementWithFunc,
	preserveDelimiters bool,
) {
	rd.FilterReplaceFull(replaceMapFunc, filterReplaceWithFunc, preserveDelimiters, true, []byte(nil))
}
