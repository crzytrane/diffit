package main

import (
	"fmt"
	"log"
  "os"
  "io"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	baseDir := "./testing/base/"
	featureDir := "./testing/feature/"
	diffDir := "./testing/diff/"

	directory := FromDirectoryOptions{
		baseDir:    baseDir,
		featureDir: featureDir,
		diffDir:    diffDir,
	}

	files, err := GetDiffsFromDirectory(directory)

	if err != nil {
		log.Fatalf("Error getting diffs from dir")
	}

	for _, diff := range files {
		_, err := DiffImage(diff, DiffOptions{Threshold: 0.1})
		if err != nil {
			log.Fatalf("Something wen't wrong")
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

    err := r.ParseMultipartForm(10 << 20) // 10 MB max
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create the destination file
    dst, err := os.Create(fmt.Sprintf("./uploads/%s", header.Filename))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy the uploaded file to the destination file
    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)

    fmt.Printf("Filename: %s, size: %s\n", header.Filename, string(header.Size))
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)
}
