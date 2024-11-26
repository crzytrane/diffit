/*
This contains the diffing logic to compare two functions
*/
package diffimage

import (
	"bufio"
	"fmt"
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

	fmt.Print("Starting diff\n")

	if (toDiff.BasePath != "" && toDiff.FeaturePath == "") || (toDiff.BasePath == "" && toDiff.FeaturePath != "") {
		return DiffResult{}, nil
	}

	file1, err := os.Open(toDiff.BasePath)
	if err != nil {
		fmt.Printf("can't open image %s %s", toDiff.BasePath, err.Error())
		return DiffResult{}, err
	}
	defer file1.Close()

	file2, err := os.Open(toDiff.FeaturePath)
	if err != nil {
		fmt.Printf("can't open image %s %s", toDiff.FeaturePath, err.Error())
		return DiffResult{}, err
	}
	defer file2.Close()

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

	fmt.Print("About to do the diff!\n")

	resultDiff := imgdiff.Diff(image1, image2, &imgdiff.Options{
		Threshold: float64(options.Threshold),
		DiffImage: true,
	})

	fmt.Print("Diff has been done!\n")

	if resultDiff.Equal {
		fmt.Print("✅ Images are equal!\n")
		result.IsEqual = true
		return result, nil
	}

	fmt.Printf("Diff written to: %s\n", toDiff.DiffPath)

	os.MkdirAll(toDiff.DiffDir, 0755)

	f, err := os.Create(toDiff.DiffPath)
	if err != nil {
		return DiffResult{}, err
	}

	writer := bufio.NewWriter(f)

	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
	enc.Encode(writer, resultDiff.Image)

	writer.Flush()

	f.Close()

	return result, nil
}
