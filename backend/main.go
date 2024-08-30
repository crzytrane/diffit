package main

import (
	"log"
)

func main() {
	baseDir := "./testing/base/"
	featureDir := "./testing/feature/"
	diffDir := "./testing/diff/"

	// collect files
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
}
