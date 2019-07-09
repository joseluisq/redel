// Package redel replaces byte occurrences between two byte delimiters.
//
// Redel provides a small interface around bufio.Scanner for replace and filter byte occurrences
// between two byte delimiters. It supports an array of byte-pair replacements
// with a map and filter closures in order to control every replacement and their values.
package redel

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"io"
)

type (
	// Redel provides an interface (around Scanner) for replace string occurrences
	// between two string delimiters.
	Redel struct {
		Reader     io.Reader
		Delimiters []Delimiter
		eof        []byte
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

// getEOFToken generates a random EOF bytes token.
func getEOFToken() []byte {
	eof := make([]byte, 7)
	_, err := rand.Read(eof)

	if err != nil {
		panic(err)
	}

	return eof
}

// New creates a new Redel instance.
func New(reader io.Reader, delimiters []Delimiter) *Redel {
	eof := getEOFToken()

	return &Redel{
		Reader:     reader,
		Delimiters: delimiters,
		eof:        eof,
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
			if from := bytes.Index(data, del.Start); from >= 0 {
				if to := bytes.Index(data[from:], del.End); to >= 0 {
					x1 := from + startLen
					x2 := from + endLen + (to - endLen)
					val := data[x1:x2]

					earlyDelimiters = append(earlyDelimiters, earlyDelimiter{
						value:      val,
						delimiter:  del,
						startIndex: x1,
						endIndex:   x2,
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
			delimiterVal := closerDelimiter.value

			if len(delimiterVal) > 0 {
				valuesData = append(valuesData, replacementData{
					delimiter: delimiter,
					value:     delimiterVal,
				})
			}

			endIndex := closerDelimiter.endIndex
			startIndex := closerDelimiter.startIndex

			return endIndex, data[0:startIndex], nil
		}

		if atEOF && len(data) > 0 {
			last := append(data[0:], rd.eof...)
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
		bytesO := scanner.Bytes()
		bytesR := make([]byte, len(bytesO))
		copy(bytesR, bytesO)

		atEOF := bytes.HasSuffix(bytesR, rd.eof)

		valueCurrent := []byte(nil)
		valueCurrentLen := len(valuesData) - 1
		valueToReplace := []byte(nil)

		var replacementData replacementData

		if valueCurrentLen >= 0 {
			replacementData = valuesData[valueCurrentLen]
			valueCurrent = append(valueCurrent, replacementData.value...)
			valueToReplace = filterFunc(valueCurrent)
		}

		delimiterData := replacementData.delimiter

		// Remove delimiters only if `preserveDelimiters` is `false`
		if !preserveDelimiters {
			// 1. Check for the first start delimiter (once)
			if !hasStartPrevDelimiter && bytes.HasSuffix(bytesR, delimiterData.Start) {
				bytesR = bytes.Replace(bytesR, delimiterData.Start, []byte(nil), 1)
				previousDelimiter = delimiterData
				hasStartPrevDelimiter = true
			}

			// 2. Next check for start and end delimiters (many times)
			if hasStartPrevDelimiter {
				hasPrevEndDelimiter := false

				// 2.1. Check for a previous end delimiter (in current data)
				if bytes.HasPrefix(bytesR, previousDelimiter.End) {
					bytesR = bytes.Replace(bytesR, previousDelimiter.End, []byte(nil), 1)
					previousDelimiter = delimiterData
					hasPrevEndDelimiter = true
				}

				// 2.2. Check for a new start delimiter (in current data)
				if hasPrevEndDelimiter && bytes.HasSuffix(bytesR, delimiterData.Start) {
					bytesR = bytes.Replace(bytesR, delimiterData.Start, []byte(nil), 1)
				}
			}
		}

		// Last process to append or not values or replacements
		if atEOF {
			bytesR = bytes.Split(bytesR, rd.eof)[0]
		} else {
			if replaceWith {
				// takes the callback value instead
				bytesR = append(bytesR, valueToReplace...)
			} else {
				// don't replace and use the value instead
				if len(valueToReplace) == 0 {
					// takes the array value instead
					bytesR = append(bytesR, valueCurrent...)
				} else {
					// otherwise use the replacement value
					bytesR = append(bytesR, replacement...)
				}
			}
		}

		replacementMapFunc(bytesR, atEOF)
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
