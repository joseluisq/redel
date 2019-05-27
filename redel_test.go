package redel

import (
	"fmt"
	"strings"
	"testing"
)

// TestStrMatch tests if a string is replaced correctly.
func TestStringMatch(t *testing.T) {
	str := "(Lorem ) ipsum dolor ( nam risus ) magna ( suscipit. ) varius ( sapien )."
	r := strings.NewReader(str)

	rep := New(r, []Delimiter{
		{Start: []byte("("), End: []byte(")")},
	})

	replacement := []byte("REPLACEMENT")
	output := ""
	expected := "Lorem ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT."

	rep.Replace(replacement, func(data []byte, atEOF bool) {
		output = output + string(data)
	})

	fmt.Println(output)

	if output != expected {
		t.Fatal("Failed to match strings!")
	}
}
