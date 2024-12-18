/*
This contains the diffing logic to compare two functions
*/
package diffimage

import (
	"bufio"
	"os"

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

/*
Compares two images and gets the differences between them
If there is a difference it will create a diff file. If images
don't exist in the base or in the feature then no diff
file will be created
*/
func DiffImage(toDiff ToDiff, options DiffOptions) (DiffResult, error) {
	result := DiffResult{
		Input:   toDiff,
		IsEqual: false,
	}

	if (toDiff.BasePath != "" && toDiff.FeaturePath == "") || (toDiff.BasePath == "" && toDiff.FeaturePath != "") {
		return DiffResult{}, nil
	}

	file1, err := os.Open(toDiff.BasePath)
	defer file1.Close()
	if err != nil {
		return DiffResult{}, err
		// log.Fatalf("can't open image %s %s", toDiff.basePath, err.Error())
	}

	file2, err := os.Open(toDiff.FeaturePath)
	defer file2.Close()
	if err != nil {
		return DiffResult{}, err
		// log.Fatalf("can't open image %s %s", toDiff.featurePath, err.Error())
	}

	image1, _, err := image.Decode(file1)
	if err != nil {
		return DiffResult{}, err
		// log.Fatalf("Error loading image1 %s\n", err)
	}

	image2, _, err := image.Decode(file2)
	if err != nil {
		return DiffResult{}, err
		// log.Fatalf("Error loading image2 %s\n", err)
	}

	resultDiff := imgdiff.Diff(image1, image2, &imgdiff.Options{
		Threshold: float64(options.Threshold),
		DiffImage: false,
	})

	if resultDiff.Equal {
		result.IsEqual = true
		return result, nil
	}

	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	// fmt.Printf("Diff written to: %s\n", toDiff.diffPath)

	os.MkdirAll(toDiff.DiffDir, 0755)

	f, err := os.Create(toDiff.DiffPath)
	if err != nil {
		return DiffResult{}, err
	}

	writer := bufio.NewWriter(f)

	enc.Encode(writer, resultDiff.Image)

	writer.Flush()

	f.Close()

	return result, nil
}
