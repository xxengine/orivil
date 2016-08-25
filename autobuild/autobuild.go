// more information see: https://github.com/orivil/watcher
package main

import (
	"gopkg.in/orivil/watcher.v0"
	"log"
	"os"
	"strings"
	"path/filepath"
)

func main() {
	// watch ".go" file
	extensions := []string{".go"}

	// handle incoming errors
	var errHandler = func(e error) {

		log.Println(e)
	}

	runner := watcher.NewAutoCommand(extensions, errHandler)

	// watch library directories
	goPath, _ := os.LookupEnv("GOPATH")
	goPaths := strings.Split(goPath, ";")
	for _, path := range goPaths {
		if path != "" {
			runner.Watch(filepath.Join(path, "src"))
		}
	}

	// build current directory
	buildFile := "."

	// run the watcher and wait for event.
	runner.RunCommand("go", "build", buildFile)
}