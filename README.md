# redel [![Build Status](https://travis-ci.org/joseluisq/redel.svg?branch=master)](https://travis-ci.org/joseluisq/redel)

> Replace string occurrences between two string delimiters.

## Install

```sh
go get github.com/joseluisq/redel
```

## Usage

__Redel__ provides an interface (around [Scanner](https://golang.org/pkg/text/scanner/)) for replace string occurrences between two string delimiters.

`Replace` function scanns and replaces string occurrences for the privided delimiters.
`Replace` requires a callback function that will be called for every successful replacement.
The callback will receive two params:
- `data` []byte (Each successful replaced byte)
- `atEOF` bool (If loop is EOF)

### String replacement

```go
package main

import (
	"github.com/joseluisq/redel"
	"fmt"
	"strings"
)

r := strings.NewReader(`Lorem ipsum dolor START nam risus END magna START suscipit. END varius START sapien END.`)

// Pass some Reader, start delimiter, end delimiter and replacement strings.
rep := redel.NewRedel(r, "START", "END", "REPLACEMENT")

// Replace function requires a callback function
// that will be called for every successful replacement.
rep.Replace(func(data []byte, atEOF bool) {
	fmt.Print(string(data))
})
// RESULT:
// Lorem ipsum dolor REPLACEMENT magna REPLACEMENT varius REPLACEMENT.⏎
```

### File replacement

```go
package main

import (
	"github.com/joseluisq/redel"
	"bufio"
	"fmt"
	"os"
)

r, err := os.Open("my_big_file.txt")

if err != nil {
	fmt.Println(err)
	os.Exit(1)
}

defer r.Close()

w, err := os.Create("my_big_file_replaced.txt")

if err != nil {
	fmt.Println(err)
	os.Exit(1)
}

defer w.Close()

var writer = bufio.NewWriter(w)

rep := redel.NewRedel(r, "START", "END", "REPLACEMENT")
rep.Replace(func(data []byte, atEOF bool) {
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

© 2017 [José Luis Quintana](http://git.io/joseluisq)
