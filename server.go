package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func (i *Interpreter) WatchAndServe(viewsDir, staticDir, port string) error {
	err := i.ParseAllTemplates(viewsDir)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Modified file:", event.Name)
					if strings.HasSuffix(event.Name, ".templ") {
						log.Println("Parsing template file:", event.Name)
						i.ParseFile(event.Name)
					} else if strings.HasPrefix(event.Name, staticDir) {
						log.Println("Static file modified:", event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	err = filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".templ") {
			log.Println("Adding watcher for template file:", path)
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Println("Adding watcher for static file:", path)
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fileServer := http.FileServer(http.Dir(staticDir))
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving static file:", r.URL.Path)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		componentName := r.URL.Query().Get("component")

		// default example
		if componentName == "" {
			componentName = "view.Home"
		}

		log.Println("Rendering component:", componentName)
		props := make(map[string]string)

		rendered, err := i.Render(componentName, props)
		if err != nil {
			log.Println("Error rendering component:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, rendered)
	})

	log.Printf("Server starting on http://localhost:%s", port)
	return http.ListenAndServe(":"+port, nil)
}

func (i *Interpreter) ParseAllTemplates(viewsDir string) error {
	return filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".templ") {
			log.Println("Parsing template file:", path)
			return i.ParseFile(path)
		}
		return nil
	})
}
