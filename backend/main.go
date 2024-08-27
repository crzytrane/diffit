package main

import (
	"bufio"
	"fmt"
	"os"

	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	"log"

	// "github.com/gofiber/fiber/v2"
	"github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

func main() {
	baseDir := "./testing/base/"
	featureDir := "./testing/feature/"

	// collect files
	directory := FromDirectoryOptions{
		baseDir:    baseDir,
		featureDir: featureDir,
	}

	files, err := GetDiffsFromDirectory(directory)

	for _, file := range files {
		fmt.Printf("Base: %s, Feature: %s\n", file.basePath, file.featurePath)
	}

	// make the diffs actually work off the diff objects
	//

  for _, diff := range files {
    result, err := DiffImage(diff, DiffOptions{Threshold: 0.1})
    if err != nil {
      log.Fatalf("Something wen't wrong")
    }
    fmt.Printf("Result: %t", result.IsEqual)
  }

	os.Exit(0)
	// diff files
	// report differences

	outputPath := "/home/mark/projects/diffit/backend/output.png"
	path1 := "/home/mark/projects/diffit/backend/tiger.png"
	path2 := "/home/mark/projects/diffit/backend/tiger-2.png"

	file1, err := os.Open(path1)
	if err != nil {
		log.Fatalf("can't open image %s %s", path1, err.Error())
	}

	file2, err := os.Open(path2)
	if err != nil {
		log.Fatalf("can't open image %s %s", path2, err.Error())
	}

	image1, _, err := image.Decode(file1)
	if err != nil {
		fmt.Printf("Error loading image1 %s\n", err)
		os.Exit(1)
	}

	image2, _, err := image.Decode(file2)
	if err != nil {
		fmt.Printf("Error loading image2 %s\n", err)
		os.Exit(1)
	}

	defer file1.Close()
	defer file2.Close()

	result := imgdiff.Diff(image1, image2, &imgdiff.Options{
		Threshold: 0.0,
		DiffImage: false,
	})

	// if result.Equal {
	//   fmt.Print("Looks like they're the same")
	//   os.Exit(0)
	// }

	enc := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	f, _ := os.Create(outputPath)

	writer := bufio.NewWriter(f)

	enc.Encode(writer, result.Image)

	writer.Flush()

	f.Close()

	// app := fiber.New(fiber.Config{
	// 	Network:      "tcp",
	// 	ServerHeader: "Fiber",
	// 	AppName:      "Backend",
	// })
	//
	// app.Get("/", func(c *fiber.Ctx) error {
	// 	{
	// 		user := MainResponse{
	// 			Id:       1,
	// 			Username: "Mark",
	// 			Password: "pass",
	// 		}
	//
	// 		return c.JSON(user)
	// 	}
	// })
	//
	// defaultPort := "9010"
	// port := cmp.Or(os.Getenv("PORT"), defaultPort)
	//
	// app.Listen(":" + port)
}
