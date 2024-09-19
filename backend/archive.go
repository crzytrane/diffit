package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func unpackArchiveFromRequest(r *http.Request) (string, error) {
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		fmt.Printf("error processing multipart form, err %s\n", err.Error())
		return "", err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("err doing FormFile\n")
		return "", err
	}
	defer file.Close()

  dst, err := os.MkdirTemp("", "extracted-")
  if err != nil {
    return "", err
  }
  // todo add this back in!
  // defer os.RemoveAll(dst)

  zipFile, err := os.CreateTemp(dst, "upload-*.zip")
  // todo add this back in!
  // defer os.Remove(zipFile.Name())
	err = os.MkdirAll("./uploads/extracted/", os.ModePerm)

  fmt.Printf("Tmp dir = %s\n", dst)

	_, err = io.Copy(zipFile, file)
	if err != nil {
		fmt.Printf("Error copying file\n")
		return "", err
	}

	archive, err := zip.OpenReader(zipFile.Name())
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

func archiveData(baseDir string, featureDir string, diffDir string) error {
  // Create a buffer to write our archive to

  // buf := new(bytes.Buffer)
  //
  // // Create a new zip archive
  // w := zip.NewWriter(buf)
  //
  // files, err := filepath.Glob(baseDir)
  // if err != nil {
  //   return err
  // }

  return nil
  // for _, file := range files {
  //   f, err := w.Create(file.Name)
  //   if err != nil {
  //     log.Fatal(err)
  //   }
  //   _, err = f.Write([]byte(file.Body))
  //   if err != nil {
  //     log.Fatal(err)
  //   }
  // }
  // err := w.Close()
  // if err != nil {
  //   log.Fatal(err)
  // }
}
