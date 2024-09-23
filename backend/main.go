package main

import (
	"cmp"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	envOrDefaultPort := cmp.Or(os.Getenv("PORT"), "3000")
  envOrDefaultPortInt, err := strconv.Atoi(envOrDefaultPort)
  if err != nil {
    panic("Couldn't convert port number, check that it is a valid value")
  }
	port := flag.Int("port", envOrDefaultPortInt, "port to run on")
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

    diffs := make([]DiffResult, len(files))
		for index, diff := range files {
			diffResult, err := DiffImage(diff, DiffOptions{Threshold: 0.1})
      fmt.Printf("Diff %d\n\t- %s\n\t- %s\n\t- %s\n", index, diffResult.Input.basePath, diffResult.Input.featurePath, diffResult.Input.diffPath)
      diffs[index] = diffResult
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

    diffTest := diffs[0].Input
    fmt.Printf("diff count %d\n\t- %s\n\t- %s\n\t- %s\n", len(diffs), diffTest.basePath, diffTest.featurePath, diffTest.diffPath)

    archiveData(baseDir, featureDir, diffDir)
	})

	fmt.Printf("Listening on port %v\n", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *port), r)

	fmt.Printf("Finished: %v", err)
}
