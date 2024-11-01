package diffimage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FromDirectoryOptions struct {
	BaseDir    string
	FeatureDir string
	DiffDir    string
}

type ToDiff struct {
	BasePath    string
	FeaturePath string
	DiffPath    string
	DiffDir     string
}

func fileExists(baseDir string, dirEntry fs.DirEntry) bool {
	filepath := baseDir + dirEntry.Name()
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

func merge[T any](arr1 *[]T, arr2 *[]T) *[]T {
	allFiles := make([]T, len(*arr1)+len(*arr2))
	copy(allFiles[:], (*arr1)[:])
	copy(allFiles[len(*arr1):], (*arr2)[:])
	return &allFiles
}

func GetDiffsFromDirectory(options FromDirectoryOptions) ([]ToDiff, error) {
	allBaseFiles, err := os.ReadDir(options.BaseDir)
	if err != nil {
		errorResult := errors.New("Error loading files from base directory")
		return []ToDiff{}, errorResult
	}
	baseFiles := filterOutDirectories(allBaseFiles)

	allFeatureFiles, err := os.ReadDir(options.FeatureDir)
	if err != nil {
		fmt.Printf("Error loading files from feature directory %s", options.FeatureDir)
		os.Exit(1)
	}
	featureFiles := filterOutDirectories(allFeatureFiles)

	results := []ToDiff{}

	allFiles := *merge(&baseFiles, &featureFiles)

	for _, file := range allFiles {
		fileExistsInBaseDir := fileExists(options.BaseDir, file)
		fileExistsInFeatureDir := fileExists(options.FeatureDir, file)
		fileExistsInBothDirs := fileExistsInBaseDir && fileExistsInFeatureDir

		if fileExistsInBothDirs {
			diff := ToDiff{
				BasePath:    options.BaseDir + file.Name(),
				FeaturePath: options.FeatureDir + file.Name(),
				DiffPath:    options.DiffDir + file.Name(),
				DiffDir:     options.DiffDir,
			}
			results = append(results, diff)
		} else if fileExistsInBaseDir {
			diff := ToDiff{
				BasePath: options.BaseDir + file.Name(),
			}
			results = append(results, diff)
		} else if fileExistsInFeatureDir {
			diff := ToDiff{
				FeaturePath: options.FeatureDir + file.Name(),
			}
			results = append(results, diff)
		}
	}

	baseFilesDirs := filterOutFiles(allBaseFiles)
	featureFilesDirs := filterOutFiles(allFeatureFiles)

	allDirs := *merge(&baseFilesDirs, &featureFilesDirs)

	for _, folder := range allDirs {
		diff, err := GetDiffsFromDirectory(FromDirectoryOptions{
			BaseDir:    options.BaseDir + folder.Name() + "/",
			FeatureDir: options.FeatureDir + folder.Name() + "/",
			DiffDir:    options.DiffDir + folder.Name() + "/",
		})
		if err != nil {
			return nil, err
		}
		results = append(results, diff...)
	}

	return dedupe(results), nil
}
