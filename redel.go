package main

import (
	"redel/cmd"
)

func main() {
	files := []string{"app", "application", "authentication", "database", "datetime.scalar"}
	// files := []string{"datetime.scalar"}

	for _, name := range files {
		cmd.Execute(name)
	}
}
