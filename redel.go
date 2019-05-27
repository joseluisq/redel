package redel

import (
	"bufio"
	"bytes"
	"fmt"
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
// - Reader io.Reader (Input reader)
// - startDelimiter string (Start string delimiter)
// - endDelimiter string (End string delimiter)
func New(reader io.Reader, delimiters []Delimiter) *Redel {
	return &Redel{
		Reader:     reader,
		Delimiters: delimiters,
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
	scanner := bufio.NewScanner(rd.Reader)
	delimiters := rd.Delimiters

	var valuesData []replacementData

	ScanByDelimiters := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		var earlyDelimiters []earlyDelimiter
		var closerDelimiter earlyDelimiter

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

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

			return closerDelimiter.endIndex, data[0:closerDelimiter.startIndex], nil
		}

		if atEOF && len(data) > 0 {
			last := append(data[0:], EOF...)
			return len(data), last, nil
		}

		return 0, nil, nil
	}

	scanner.Split(ScanByDelimiters)

	counterDelimiterStart := false

	// Scan every token based on current split function
	for scanner.Scan() {
		bytesR := append([]byte(nil), scanner.Bytes()...)
		atEOF := bytes.HasSuffix(bytesR, EOF)

		// Checks for a valid value
		value := []byte(nil)
		valuesLen := len(valuesData) - 1
		valueToReplace := []byte(nil)

		var replacementData replacementData

		if !atEOF && valuesLen >= 0 {
			replacementData = valuesData[valuesLen]
			value = append([]byte(nil), replacementData.value...)
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

		delimiter := replacementData.delimiter

		// Preserve or remove delimiters
		if !preserveDelimiters {
			// TODO: Replace delimiters propertly
			fmt.Println("DATA:", string(bytesW))

			if bytes.Index(bytesW, delimiter.Start) >= 0 {
				fmt.Println("VALUE:", string(value))
				fmt.Println("--")
				bytesW = bytes.Replace(bytesW, delimiter.Start, []byte(nil), 1)
				counterDelimiterStart = true
			}

			if counterDelimiterStart && bytes.Index(bytesW, delimiter.End) >= 0 {
				bytesW = bytes.Replace(bytesW, delimiter.End, []byte(nil), 1)
				counterDelimiterStart = false
				fmt.Println("END:", string(delimiter.End))
				fmt.Println("------------------------------------------------")
				fmt.Println("")
			}
		}

		replacementMapFunc(bytesW, atEOF)
	}
}

// Replace function replaces every occurrence with a custom replacement token.
func (rd *Redel) Replace(replacement []byte, replacementMapFunc ReplacementMapFunc) {
	rd.ReplaceFilterWith(replacementMapFunc, func(value []byte) []byte {
		return replacement
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
