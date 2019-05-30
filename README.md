# redel [![Build Status](https://travis-ci.org/joseluisq/redel.svg?branch=master)](https://travis-ci.org/joseluisq/redel)

> Replace byte occurrences between two byte delimiters.

__Redel__ provides an small interface around [Scanner](https://golang.org/pkg/text/scanner/) for replace and filter byte occurrences between two byte delimiters. It supports an array of byte-pairs replacements with a map and filter callbacks as params in order to control every replacement.

## Install

```sh
go get github.com/joseluisq/redel
```

## API

### New

It creates a new `Redel` instance.

```go
func New(reader io.Reader, delimiters []Delimiter) *Redel
```

### Replace

`Replace` function replaces every occurrence with a custom replacement token.

```go
func Replace(replacement []byte, mapFunc ReplacementMapFunc)
```

### ReplaceFilter

`ReplaceFilter` function scans and replaces byte occurrences filtering every replacement value via a bool callback.

```go
func ReplaceFilter(replacement []byte, mapFunc ReplacementMapFunc, filterFunc FilterValueFunc, preserveDelimiters bool)
```

### ReplaceFilterWith

`ReplaceFilterWith` function scans and replaces byte occurrences filtering every replacement value via a custom replacement callback.

```go
func ReplaceFilterWith(mapFunc ReplacementMapFunc, filterReplaceFunc FilterValueReplaceFunc, preserveDelimiters bool)
```

## Usage

### String replacement

```go
package main

import (
	"fmt"
	"strings"

	"github.com/joseluisq/redel"
)

reader := strings.NewReader(`Lorem ipsum dolor START nam risus END magna START suscipit. END varius START sapien END.`)

// Pass some Reader and bytes delimiters (`Start` and `End`)
r := redel.New(reader, []Delimiter{
	{Start: []byte("START"), End: []byte("END")},
})

// Replace function requires a callback function
// that will be called for every successful replacement.
r.Replace([]byte("REPLACEMENT"), func(data []byte, atEOF bool) {
	fmt.Print(string(data))
})
// RESULT:
// Lorem ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT.⏎
```

### File replacement

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/joseluisq/redel"
)

reader, err := os.Open("my_big_file.txt")

if err != nil {
	fmt.Println(err)
	os.Exit(1)
}

defer reader.Close()

writer, err := os.Create("my_big_file_replaced.txt")

if err != nil {
	fmt.Println(err)
	os.Exit(1)
}

defer writer.Close()

var writer = bufio.NewWriter(writer)

// Use Redel API
r := redel.New(reader, []Delimiter{
	{Start: []byte("START"), End: []byte("END")},
})
r.Replace([]byte("REPLACEMENT"), (func(data []byte, atEOF bool) {
	_, err := writer.Write(data)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
})

writer.Flush()
```

## Contributions

[Pull requests](https://github.com/joseluisq/redel/pulls) and [issues](https://github.com/joseluisq/redel/issues) are very appreciated.

## License
MIT license

© 2017-present [Jose Quintana](http://git.io/joseluisq)
