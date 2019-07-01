// Replace byte occurrences between two byte delimiters.
//
// Redel provides a small interface around bufio.Scanner for replace and filter byte occurrences
// between two byte delimiters. It supports an array of byte-pair replacements
// with a map and filter closures in order to control every replacement and their values.
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
		Reader     io.Reader
		Delimiters []Delimiter
	}

	// Delimiter defines a replacement delimiters structure
	Delimiter struct {
		Start []byte
		End   []byte
	}

	// replacementData interface contains intern replacing info.
	replacementData struct {
		delimiter Delimiter
		value     []byte
	}

	// earlyDelimiter defines a found delimiter
	earlyDelimiter struct {
		value      []byte
		delimiter  Delimiter
		startIndex int
		endIndex   int
	}

	// ReplacementMapFunc defines a map function that will be called for every scan splitted token.
	ReplacementMapFunc func(data []byte, atEOF bool)

	// FilterValueFunc defines a filter function that will be called per replacement
	// which supports a return `bool` value to apply the replacement or not.
	FilterValueFunc func(matchValue []byte) bool

	// FilterValueReplaceFunc defines a filter function that will be called per replacement
	// which supports a return `[]byte` value to customize the replacement value.
	FilterValueReplaceFunc func(matchValue []byte) []byte
)

var (
	// EOF is an byte intended to determine the EOF in scanning.
	EOF = []byte("eof")
)

// New creates a new Redel instance.
func New(reader io.Reader, delimiters []Delimiter) *Redel {
	return &Redel{
		Reader:     reader,
		Delimiters: delimiters,
	}
}

// replaceFilterFunc is the API function which scans and replace bytes supporting different options.
// It's used by API's replace functions.
func (rd *Redel) replaceFilterFunc(
	replacementMapFunc ReplacementMapFunc,
	filterFunc FilterValueReplaceFunc,
	preserveDelimiters bool,
	replaceWith bool,
	replacement []byte,
) {
	scanner := bufio.NewScanner(rd.Reader)
	delimiters := rd.Delimiters

	var valuesData []replacementData

	ScanByDelimiters := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		var earlyDelimiters []earlyDelimiter
		var closerDelimiter earlyDelimiter

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// iterate array of delimiters
		for _, del := range delimiters {
			startLen := len(del.Start)
			endLen := len(del.End)

			if startLen <= 0 || endLen <= 0 {
				continue
			}

			// store every found delimiter
			if from := bytes.Index(data, []byte(del.Start)); from >= 0 {
				if to := bytes.Index(data[from:], []byte(del.End)); to >= 0 {
					a := from + startLen
					b := from + endLen + (to - endLen)
					val := data[a:b]

					earlyDelimiters = append(earlyDelimiters, earlyDelimiter{
						value:      val,
						delimiter:  del,
						startIndex: a,
						endIndex:   b,
					})
				}
			}
		}

		if len(earlyDelimiters) > 0 {
			// Determine the closer delimiter
			for i, del := range earlyDelimiters {
				if i == 0 || del.startIndex < closerDelimiter.startIndex {
					closerDelimiter = del
				}
			}

			// Assign and check the closer delimiter
			delimiter := closerDelimiter.delimiter
			val := closerDelimiter.value

			if len(val) > 0 {
				valuesData = append(valuesData, replacementData{
					delimiter: delimiter,
					value:     val,
				})
			}

			endIndex := closerDelimiter.endIndex
			startIndex := closerDelimiter.startIndex

			return endIndex, data[0:startIndex], nil
		}

		if atEOF && len(data) > 0 {
			last := append(data[0:], EOF...)
			return len(data), last, nil
		}

		return 0, nil, nil
	}

	scanner.Split(ScanByDelimiters)

	// Variables to control delimiters checking
	hasStartPrevDelimiter := false
	var previousDelimiter Delimiter

	// Scan every token based on current split function
	for scanner.Scan() {
		bytesR := append([]byte(nil), scanner.Bytes()...)
		atEOF := bytes.HasSuffix(bytesR, EOF)

		// Checks for a valid value
		value := []byte(nil)
		valuesLen := len(valuesData) - 1
		valueToReplace := []byte(nil)

		var replacementData replacementData

		if valuesLen >= 0 {
			replacementData = valuesData[valuesLen]
			value = append([]byte(nil), replacementData.value...)
			valueToReplace = filterFunc(value)
		}

		bytesW := bytesR
		delimiterData := replacementData.delimiter

		// Remove delimiters only if `preserveDelimiters` is `false`
		if !preserveDelimiters {
			// 1. Check for the first start delimiter (once)
			if !hasStartPrevDelimiter && bytes.HasSuffix(bytesW, delimiterData.Start) {
				bytesW = bytes.Replace(bytesW, delimiterData.Start, []byte(nil), 1)
				previousDelimiter = delimiterData
				hasStartPrevDelimiter = true
			}

			// 2. Next check for start and end delimiters (many times)
			if hasStartPrevDelimiter {
				hasPrevEndDelimiter := false

				// 2.1. Check for a previous end delimiter (in current data)
				if bytes.HasPrefix(bytesW, previousDelimiter.End) {
					bytesW = bytes.Replace(bytesW, previousDelimiter.End, []byte(nil), 1)
					previousDelimiter = delimiterData
					hasPrevEndDelimiter = true
				}

				// 2.2. Check for a new start delimiter (in current data)
				if hasPrevEndDelimiter && bytes.HasSuffix(bytesW, delimiterData.Start) {
					bytesW = bytes.Replace(bytesW, delimiterData.Start, []byte(nil), 1)
					hasPrevEndDelimiter = false
				}
			}
		}

		// Last process to append or not values or replacements
		if atEOF {
			bytesW = bytes.Split(bytesW, EOF)[0]
		} else {
			if replaceWith {
				// takes the callback value instead
				bytesW = append(bytesW, valueToReplace...)
			} else {
				// don't replace and use the value instead
				if len(valueToReplace) == 0 {
					// takes the array value instead
					bytesW = append(bytesW, value...)
				} else {
					// otherwise use the replacement value
					bytesW = append(bytesW, replacement...)
				}
			}
		}

		replacementMapFunc(bytesW, atEOF)
	}
}

// Replace function replaces every occurrence with a custom replacement token.
func (rd *Redel) Replace(replacement []byte, mapFunc ReplacementMapFunc) {
	rd.replaceFilterFunc(mapFunc, func(value []byte) []byte {
		return value
	}, false, false, replacement)
}

// ReplaceFilter function scans and replaces byte occurrences filtering every replacement value via a bool callback.
func (rd *Redel) ReplaceFilter(
	replacement []byte,
	mapFunc ReplacementMapFunc,
	filterFunc FilterValueFunc,
	preserveDelimiters bool,
) {
	rd.replaceFilterFunc(mapFunc, func(matchValue []byte) []byte {
		result := []byte(nil)

		ok := filterFunc(matchValue)

		if ok {
			result = []byte("1")
		}

		return result
	}, preserveDelimiters, false, replacement)
}

// ReplaceFilterWith function scans and replaces byte occurrences via a custom replacement callback.
func (rd *Redel) ReplaceFilterWith(
	mapFunc ReplacementMapFunc,
	filterReplaceFunc FilterValueReplaceFunc,
	preserveDelimiters bool,
) {
	rd.replaceFilterFunc(mapFunc, filterReplaceFunc, preserveDelimiters, true, []byte(nil))
}
