package main

import (
	"fmt"

	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func main() {

	port := flag.Int("port", 3000, "port to run on")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!\n"))
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		baseFilePath, err := unpackArchiveFromRequest(r)

		baseDir := fmt.Sprintf("%s/base/", baseFilePath)
		featureDir := fmt.Sprintf("%s/feature/", baseFilePath)
		diffDir := fmt.Sprintf("%s/diff/", baseFilePath)

		fmt.Printf("baseFilePath is: %s\n", baseFilePath)

		fromDirectoryOptions := FromDirectoryOptions{
			baseDir:    baseDir,
			featureDir: featureDir,
			diffDir:    diffDir,
		}

		if err != nil {
			// todo validate this is correct
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files, err := GetDiffsFromDirectory(fromDirectoryOptions)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, diff := range files {
			_, err := DiffImage(diff, DiffOptions{Threshold: 0.1})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	})

	fmt.Printf("Listening on port %v\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), r)

	fmt.Printf("Finished: %v", err)
}
