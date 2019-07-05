package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"

	redel "redel/pkg"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(name string) {
	r, err := os.Open(".tmp/" + name + ".js")

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer r.Close()

	w, err := os.Create("./.tmp/dist/" + name + ".repl.js")

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer w.Close()

	var writer = bufio.NewWriter(w)

	replaceFunc := func(data []byte, atEOF bool) {
		_, err := writer.Write(data)

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	filterFunc := func(matchValue []byte) []byte {
		pattern := []byte("~/")

		val := string(matchValue)
		pat := string(pattern)
		fmt.Println("Current value:", val)
		fmt.Println("Replacement:", pat)

		if bytes.HasPrefix(matchValue, pattern) {
			repl := bytes.Replace(matchValue, pattern, []byte("./"), 1)
			vrepl := string(repl)

			fmt.Println("Has placeholder to replace:", vrepl)
			fmt.Println("")
			return repl
		}

		fmt.Println("")

		return matchValue
	}

	rep := redel.New(r, []redel.Delimiter{
		{Start: []byte("require(\""), End: []byte("\")")},
	})

	rep.ReplaceFilterWith(replaceFunc, filterFunc, true)

	writer.Flush()
}
