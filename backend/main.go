package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	// "log"

	"flag"

	"strings"

	"archive/zip"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	port := flag.Int("port", 3000, "port to run on")
	flag.Parse()

	// baseDir := "./testing/base/"
	// featureDir := "./testing/feature/"
	// diffDir := "./testing/diff/"
	//
	// directory := FromDirectoryOptions{
	// 	baseDir:    baseDir,
	// 	featureDir: featureDir,
	// 	diffDir:    diffDir,
	// }
	//
	// files, err := GetDiffsFromDirectory(directory)
	//
	// if err != nil {
	// 	log.Fatalf("Error getting diffs from dir")
	// }
	//
	// for _, diff := range files {
	// 	_, err := DiffImage(diff, DiffOptions{Threshold: 0.1})
	// 	if err != nil {
	// 		log.Fatalf("Something wen't wrong")
	// 	}
	// }

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!\n"))
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		baseFilePath, err := decodeArchiveFromRequest(r)

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

func decodeArchiveFromRequest(r *http.Request) (string, error) {
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		fmt.Printf("error processing multipart form, err %s\n", err.Error())
		return "", err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("err doing FormFile\n")
		return "", err
	}
	defer file.Close()

  err = os.MkdirAll("./uploads/extracted/", os.ModePerm)

	zipPath := fmt.Sprintf("./uploads/%s", header.Filename)
	dst := fmt.Sprintf("./uploads/extracted/%s/", strings.Replace(header.Filename, ".zip", "", -1))

  fmt.Printf("Return value will be %s\n", dst)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		fmt.Printf("Error creating file\n")
		return "", err
	}
	defer zipFile.Close()

	_, err = io.Copy(zipFile, file)
	if err != nil {
		fmt.Printf("Error copying file\n")
		return "", err
	}

	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		fmt.Printf("Failed to open zip\n")
		return "", err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		// fmt.Println("Unzipping file", filePath)

		// Check for Zip Slip vulnerability
		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return "", errors.New("Invalid file path")
		}

		// Create directories if the entry is a directory
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// Create the destination directory if necessary
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", err
		}

		// Open the file in the ZIP archive
		fileInArchive, err := f.Open()
		if err != nil {
			return "", err
		}
		defer fileInArchive.Close()

		// Create a new file in the destination directory
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}
		defer dstFile.Close()

		// Copy the contents of the file
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return "", err
		}

		// todo probably remove these when this is working as intended
		// fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)

		// fmt.Printf("Filename: %s, size: %s\n", header.Filename, string(header.Size))
		// todo return something better later on
	}

	return dst, nil
}
