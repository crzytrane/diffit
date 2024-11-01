package main

import (
	"cmp"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/crzytrane/diffit/internal/archive"
	"github.com/crzytrane/diffit/internal/diffimage"
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

	r.Post("/api/files", func(w http.ResponseWriter, r *http.Request) {
		// get files written locally(or do I just keep them in memory 🤔)
		err := r.ParseMultipartForm(32 << 20) // 32 MB max
		if err != nil {
			fmt.Printf("error processing multipart form, err %s\n", err.Error())
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("err doing FormFile\n")
			return
		}
		defer file.Close()
	})

	r.Post("/api/archive", func(w http.ResponseWriter, r *http.Request) {
		baseFilePath, err := archive.UnpackArchiveFromRequest(r)

		baseDir := fmt.Sprintf("%s/base/", baseFilePath)
		featureDir := fmt.Sprintf("%s/feature/", baseFilePath)
		diffDir := fmt.Sprintf("%s/diff/", baseFilePath)

		fmt.Printf("baseFilePath is: %s\n", baseFilePath)

		fromDirectoryOptions := diffimage.FromDirectoryOptions{
			BaseDir:    baseDir,
			FeatureDir: featureDir,
			DiffDir:    diffDir,
		}

		if err != nil {
			// todo validate this is correct
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files, err := diffimage.GetDiffsFromDirectory(fromDirectoryOptions)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		diffs := make([]diffimage.DiffResult, len(files))
		for index, diff := range files {
			diffResult, err := diffimage.DiffImage(diff, diffimage.DiffOptions{Threshold: 0.1})
			fmt.Printf("Diff %d\n\t- %s\n\t- %s\n\t- %s\n", index, diffResult.Input.BasePath, diffResult.Input.FeaturePath, diffResult.Input.DiffPath)
			diffs[index] = diffResult
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		diffTest := diffs[0].Input
		fmt.Printf("diff count %d\n\t- %s\n\t- %s\n\t- %s\n", len(diffs), diffTest.BasePath, diffTest.FeaturePath, diffTest.DiffPath)

		archive.ArchiveData(baseDir, featureDir, diffDir)

		http.Redirect(w, r, "http://localhost:5173/", http.StatusFound)
	})

	fmt.Printf("Listening on port %v\n", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", *port), r)

	fmt.Printf("Finished: %v", err)
}
