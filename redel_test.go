package redel

import (
	"strings"
	"testing"
)

// TestStrMatch tests if a string is replaced correctly.
func TestStringMatch(t *testing.T) {
	r := strings.NewReader(`Lorem ipsum dolor START nam risus END magna START suscipit. END varius START sapien END.`)

	rep := NewRedel(r, "START", "END", "REPLACEMENT")

	var output = ""
	var expected = "Lorem ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT."

	rep.Replace(func(data []byte, atEOF bool) {
		output = output + string(data)
	})

	if output != expected {
		t.Fatal("Failed to match strings!")
	}
}
