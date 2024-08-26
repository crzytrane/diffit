package main

import (
	"log"
	"os"

	"bufio"

	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	"github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

type DiffResult struct {
	IsEqual bool
	Input   ToDiff
}

type DiffOptions struct {
	Threshold float64
}

func DiffImage(toDiff ToDiff, options DiffOptions) (DiffResult, error) {
	outputPath := "/home/mark/projects/diffit/backend/output.png"
	result := DiffResult{
		Input:   toDiff,
		IsEqual: false,
	}

	file1, err := os.Open(toDiff.basePath)
	defer file1.Close()
	if err != nil {
		// TODO change Fatalf to proper error returns, do this for all of them
		log.Fatalf("can't open image %s %s", toDiff.basePath, err.Error())
	}

	file2, err := os.Open(toDiff.featurePath)
	defer file2.Close()
	if err != nil {
		log.Fatalf("can't open image %s %s", toDiff.featurePath, err.Error())
	}

	image1, _, err := image.Decode(file1)
	if err != nil {
		log.Fatalf("Error loading image1 %s\n", err)
	}

	image2, _, err := image.Decode(file2)
	if err != nil {
		log.Fatalf("Error loading image2 %s\n", err)
	}

	resultDiff := imgdiff.Diff(image1, image2, &imgdiff.Options{
		Threshold: float64(options.Threshold),
		DiffImage: false,
	})

	if resultDiff.Equal {
		result.IsEqual = true
	}

	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	f, _ := os.Create(outputPath)

	writer := bufio.NewWriter(f)

	enc.Encode(writer, resultDiff.Image)

	writer.Flush()

	f.Close()

	return result, nil
}
