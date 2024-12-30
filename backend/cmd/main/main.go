package main

import (
	"cmp"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/crzytrane/diffit/internal/archive"
	"github.com/crzytrane/diffit/internal/diffimage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"bufio"

	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	"github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

func body(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Write([]byte("Hello world!\n"))
}

func main() {
	envOrDefaultPort := cmp.Or(os.Getenv("PORT"), "8080")
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

	r.Get("/", body)
	r.Get("/health", body)
	r.Get("/api/", body)
	r.Get("/api/health", body)

	r.Post("/api/files", func(w http.ResponseWriter, r *http.Request) {
		// get files written locally(or do I just keep them in memory 🤔)
		err := r.ParseMultipartForm(32 << 20) // 32 MB max
		if err != nil {
			fmt.Printf("error processing multipart form, err %s\n", err.Error())
			return
		}

		baseUpload, _, err := r.FormFile("file-base")
		if err != nil {
			fmt.Printf("err uploading base file\n")
			return
		}
		defer baseUpload.Close()

		otherUpload, _, err := r.FormFile("file-other")
		if err != nil {
			fmt.Printf("err uploading feature file\n")
			return
		}
		defer otherUpload.Close()

		dst, err := os.MkdirTemp("", "extracted-")
		if err != nil {
			return
		}

		baseFile, err := os.CreateTemp(dst, "base-*.png")
		if err != nil {
			return
		}
		defer baseFile.Close()
		otherFile, err := os.CreateTemp(dst, "other-*.png")
		if err != nil {
			return
		}
		defer otherFile.Close()
		diffFile, err := os.CreateTemp(dst, "diff-*.png")
		if err != nil {
			return
		}
		defer diffFile.Close()

		_, err = io.Copy(baseFile, baseUpload)
		if err != nil {
			fmt.Print("failed to copy base file\n")
			return
		}
		_, err = io.Copy(otherFile, otherUpload)
		if err != nil {
			fmt.Print("failed to copy other file\n")
			return
		}

		_, err = baseFile.Seek(0, io.SeekStart)
		if err != nil {
			fmt.Print("failed to seek files\n")
			return
		}
		_, err = otherFile.Seek(0, io.SeekStart)
		if err != nil {
			fmt.Print("failed to seek files\n")
			return
		}

		image1, _, err := image.Decode(baseFile)
		if err != nil {
			fmt.Print("failed to decode base file\n")
			return
		}

		image2, _, err := image.Decode(otherFile)
		if err != nil {
			fmt.Print("failed to decode other file\n")
			return
		}

		threshold := float64(0.1)
		resultDiff := *imgdiff.Diff(image1, image2, &imgdiff.Options{
			Threshold: threshold,
			DiffImage: false,
		})

		writer := bufio.NewWriter(diffFile)

		enc := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		w.WriteHeader(http.StatusOK)

		err = enc.Encode(w, resultDiff.Image)

		if err != nil {
			fmt.Print("failed to encode diff file\n")
		}
		err = writer.Flush()
		if err != nil {
			fmt.Print("failed to flush diff file\n")
		}

		log.Printf("diff path is %s, they are the same %v", diffFile.Name(), resultDiff.Equal)

		id := r.Header.Get("RequestId")
		test := fmt.Sprintf("/?diff=%s", id)
		fmt.Printf("RequestId %s", test)

		// w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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
