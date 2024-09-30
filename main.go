package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	interpreter := NewInterpreter()
	err := filepath.Walk("./views", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".templ") {
			return interpreter.ParseFile(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = interpreter.WatchAndServe("./views", "./static", "8080")
	if err != nil {
		log.Fatal(err)
	}
}
