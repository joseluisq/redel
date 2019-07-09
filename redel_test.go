package redel

import (
	"bytes"
	"strings"
	"testing"
)

const STR = "(Lorem ( ) ipsum dolor [ nam risus ] magna ( suscipit. ) varius { sapien }."

var delimiters = []Delimiter{
	{Start: []byte("["), End: []byte("]")},
	{Start: []byte("{"), End: []byte("}")},
	{Start: []byte("("), End: []byte(")")},
}

func TestReplaceString(t *testing.T) {
	r := strings.NewReader(STR)

	rep := New(r, delimiters)

	expectedStr := "REPLACEMENT ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT."
	replacement := []byte("REPLACEMENT")
	output := ""

	rep.Replace(replacement, func(data []byte, atEOF bool) {
		output = output + string(data)
	})

	if output != expectedStr {
		t.Fatal("1. (Replace) Failed to match strings!")
	}
}

func TestReplaceFilterString(t *testing.T) {
	r := strings.NewReader(STR)

	rep := New(r, delimiters)

	expectedStr := "REPLACEMENT ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT."
	replacement := []byte("REPLACEMENT")

	output := ""

	filterFunc := func(matchValue []byte) bool {
		return true
	}

	rep.ReplaceFilter(replacement, func(data []byte, atEOF bool) {
		output = output + string(data)
	}, filterFunc, false)

	if output != expectedStr {
		t.Fatal("2. (ReplaceFilter with no preserve delimiters) Failed to match strings!")
	}
}

func TestReplaceFilterPreserveString(t *testing.T) {
	r := strings.NewReader(STR)

	rep := New(r, delimiters)

	expectedStr := "(REPLACEMENT) ipsum dolor [REPLACEMENT] magna (REPLACEMENT) varius {REPLACEMENT}."
	replacement := []byte("REPLACEMENT")

	output := ""

	filterFunc := func(matchValue []byte) bool {
		return true
	}

	rep.ReplaceFilter(replacement, func(data []byte, atEOF bool) {
		output = output + string(data)
	}, filterFunc, true)

	if output != expectedStr {
		t.Fatal("3. (ReplaceFilter + preserve delimiters) Failed to match strings!")
	}
}
func TestReplaceFilterWithString(t *testing.T) {
	r := strings.NewReader(STR)

	rep := New(r, delimiters)

	expectedStr := "Lorem (  ipsum dolor CUSTOM magna  suscipit.  varius CUSTOM."
	replaceWithThis := []byte("CUSTOM")

	output := ""

	filterFunc := func(matchValue []byte) []byte {
		if bytes.Equal(matchValue, []byte(" sapien ")) || bytes.Equal(matchValue, []byte(" nam risus ")) {
			return replaceWithThis
		}

		return matchValue
	}

	rep.ReplaceFilterWith(func(data []byte, atEOF bool) {
		output = output + string(data)
	}, filterFunc, false)

	if output != expectedStr {
		t.Fatal("4. (ReplaceFilterWith + no preserve delimiters) Failed to match strings!")
	}
}
func TestReplaceFilterWithPreserveString(t *testing.T) {
	r := strings.NewReader(STR)

	rep := New(r, delimiters)

	expectedStr := "(Lorem ( ) ipsum dolor [ nam risus ] magna ( suscipit. ) varius {CUSTOM}."
	replaceWithThis := []byte("CUSTOM")
	hasThisValue := []byte(" sapien ")

	output := ""

	filterFunc := func(matchValue []byte) []byte {
		if bytes.Equal(matchValue, hasThisValue) {
			return replaceWithThis
		}

		return matchValue
	}

	rep.ReplaceFilterWith(func(data []byte, atEOF bool) {
		output = output + string(data)
	}, filterFunc, true)

	if output != expectedStr {
		t.Fatal("5. (ReplaceFilterWith + preserve delimiters) Failed to match strings!")
	}
}
