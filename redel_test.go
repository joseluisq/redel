package redel

import (
	"strings"
	"testing"
)

// TestStrMatch tests if a string is replaced correctly.
func TestReplaceString(t *testing.T) {
	str := "(Lorem ( ) ipsum dolor ( nam risus ) magna ( suscipit. ) varius ( sapien )."

	r := strings.NewReader(str)

	rep := New(r, []Delimiter{
		{Start: []byte("("), End: []byte(")")},
	})

	expectedStr := "REPLACEMENT ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT."
	replacement := []byte("REPLACEMENT")
	output := ""

	rep.Replace(replacement, func(data []byte, atEOF bool) {
		output = output + string(data)
	})

	if output != expectedStr {
		t.Fatal("(Replace) Failed to match strings!")
	}
}

func TestReplaceFilterString(t *testing.T) {
	str := "(Lorem ( ) ipsum dolor ( nam risus ) magna ( suscipit. ) varius ( sapien )."

	r := strings.NewReader(str)

	rep := New(r, []Delimiter{
		{Start: []byte("("), End: []byte(")")},
	})

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
		t.Fatal("(ReplaceFilter) Failed to match strings!")
	}
}
