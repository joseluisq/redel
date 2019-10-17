# Redel [![Build Status](https://travis-ci.com/joseluisq/redel.svg?branch=master)](https://travis-ci.com/joseluisq/redel) [![codecov](https://codecov.io/gh/joseluisq/redel/branch/master/graph/badge.svg)](https://codecov.io/gh/joseluisq/redel) [![Go Report Card](https://goreportcard.com/badge/github.com/joseluisq/redel)](https://goreportcard.com/report/github.com/joseluisq/redel) [![GoDoc](https://godoc.org/github.com/joseluisq/redel?status.svg)](https://godoc.org/github.com/joseluisq/redel)

> Replace byte occurrences between two byte delimiters.

__Redel__ provides a small interface around [bufio.Scanner](https://golang.org/pkg/bufio/#Scanner) for replace and filter byte occurrences between two byte delimiters. It supports an array of byte-pair replacements with a map and filter closures in order to control every replacement and their values.

## Supported Go versions

- 1.10.3+
- 1.11+

💡 For older versions, please use the latest `v2` tag.

## Usage

### String replacement

```go
package main

import (
	"fmt"
	"strings"

	"github.com/joseluisq/redel/v3"
)

func main() {
	// 1. Configure a Reader.
	str := "Lorem ipsum dolor START nam risus END magna START suscipit. END varius START sapien END."
	reader := strings.NewReader(str)

	// 2. Intance Redel using a Reader and an array of byte delimiters.
	rd := redel.New(reader, []redel.Delimiter{
		// 2.1 Define here the byte delimiters which ones should be applied
		{Start: []byte("START"), End: []byte("END")},

		// Note that this byte-pair is not present in our example,
		// so it will be not applied.
		{Start: []byte("BEGIN"), End: []byte("END")},
	})

	// 3. Finally, define a byte replacement and then replace occurrences.
	//    Replace supports a closure which will be called for every scan-splitted token.
	replacement := []byte("REPLACEMENT")
	rd.Replace(replacement, func(data []byte, atEOF bool) {
		// print out only for demonstration
		fmt.Print(string(data))
	})

	// RESULT:
	// Lorem ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT.⏎
}
```

### File replacement

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/joseluisq/redel/v3"
)

func main() {
	// 1. Configure a Reader.
	reader, err := os.Open("my_big_file.txt")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer reader.Close()

	f, err := os.Create("my_big_file_replaced.txt")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer f.Close()

	var writer = bufio.NewWriter(f)

	// 2. Intance Redel using a Reader and an array of byte delimiters.
	replacement := []byte("REPLACEMENT")
	rd := redel.New(reader, []redel.Delimiter{
		// 2.1 Define here the byte delimiters which ones should be applied
		{Start: []byte("START"), End: []byte("END")},
		{Start: []byte("BEGIN"), End: []byte("END")},
	})

	// 3. Finally, define a byte replacement, replace occurrences and
	//    save every scan-splitted token to the file.
	rd.Replace(replacement, func(data []byte, atEOF bool) {
		_, err := writer.Write(data)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	writer.Flush()
}
```

More API examples can be found in [redel_test.go](./redel_test.go) file.

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

`ReplaceFilterWith` function scans and replaces byte occurrences filtering every matched replacement value and supporting a value callback in order to customize those values.

```go
func ReplaceFilterWith(mapFunc ReplacementMapFunc, filterReplaceFunc FilterValueReplaceFunc, preserveDelimiters bool)
```

## Contributions

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in current work by you, as defined in the Apache-2.0 license, shall be dual licensed as described below, without any additional terms or conditions.

Feel free to send some [Pull request](https://github.com/joseluisq/redel/pulls) or [issue](https://github.com/joseluisq/redel/issues).

## License

This work is primarily distributed under the terms of both the [MIT license](LICENSE-MIT) and the [Apache License (Version 2.0)](LICENSE-APACHE).

© 2017-present [Jose Quintana](http://git.io/joseluisq)
