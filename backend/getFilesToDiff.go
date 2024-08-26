package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FromDirectoryOptions struct {
	baseDir    string
	featureDir string
}

type ToDiff struct {
	basePath    string
	featurePath string
}

func fileExists(baseDir string, dirEntry fs.DirEntry) bool {
	filepath := baseDir + dirEntry.Name()
	// fmt.Printf("fileExits baseDir = %s, dirEntry.Name() = %s fullPath = %s \n", baseDir, dirEntry.Name(), filepath)
	_, err := os.Stat(filepath)
	return !errors.Is(err, os.ErrNotExist)
}

func filterOutDirectories(files []fs.DirEntry) []fs.DirEntry {
	filteredFiles := []fs.DirEntry{}
	for _, file := range files {
		if !file.IsDir() {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles
}

func filterOutFiles(files []fs.DirEntry) []fs.DirEntry {
	filteredDirs := []fs.DirEntry{}
	for _, file := range files {
		if file.IsDir() {
			filteredDirs = append(filteredDirs, file)
		}
	}
	return filteredDirs
}

func dedupe(slice []ToDiff) []ToDiff {
	keys := make(map[ToDiff]bool)
	list := []ToDiff{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Merge[T any](arr1 *[]T, arr2 *[]T) *[]T {
	allFiles := make([]T, len(*arr1)+len(*arr2))
	copy(allFiles[:], (*arr1)[:])
	copy(allFiles[len(*arr1):], (*arr2)[:])
	return &allFiles
}

func GetDiffsFromDirectory(options FromDirectoryOptions) ([]ToDiff, error) {
	allBaseFiles, err := os.ReadDir(options.baseDir)
	if err != nil {
		errorResult := errors.New("Error loading files from base directory")
		return []ToDiff{}, errorResult
	}
	baseFiles := filterOutDirectories(allBaseFiles)

	allFeatureFiles, err := os.ReadDir(options.featureDir)
	if err != nil {
		fmt.Printf("Error loading files from feature directory %s", options.featureDir)
		os.Exit(1)
	}
	featureFiles := filterOutDirectories(allFeatureFiles)

	results := []ToDiff{}

	allFiles := *Merge(&baseFiles, &featureFiles)

	for _, file := range allFiles {
		fileExistsInBaseDir := fileExists(options.baseDir, file)
		fileExistsInFeatureDir := fileExists(options.featureDir, file)
		fileExistsInBothDirs := fileExistsInBaseDir && fileExistsInFeatureDir

		if fileExistsInBothDirs {
			diff := ToDiff{
				basePath:    options.baseDir + file.Name(),
				featurePath: options.featureDir + file.Name(),
			}
			results = append(results, diff)
		} else if fileExistsInBaseDir {
			diff := ToDiff{
				basePath: options.baseDir + file.Name(),
			}
			results = append(results, diff)
		} else if fileExistsInFeatureDir {
			diff := ToDiff{
				featurePath: options.featureDir + file.Name(),
			}
			results = append(results, diff)
		}
	}

	baseFilesDirs := filterOutFiles(allBaseFiles)
	featureFilesDirs := filterOutFiles(allFeatureFiles)

	allDirs := *Merge(&baseFilesDirs, &featureFilesDirs)

	for _, folder := range allDirs {
		diff, err := GetDiffsFromDirectory(FromDirectoryOptions{
			baseDir:    options.baseDir + folder.Name() + "/",
			featureDir: options.featureDir + folder.Name() + "/",
		})
		if err != nil {
			return nil, err
		}
		results = append(results, diff...)
	}

	return dedupe(results), nil
}
