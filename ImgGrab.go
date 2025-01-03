package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"image"
)

func ImgGrab() {
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", err)
		return
	}
	defer webcam.Close()
	window := gocv.NewWindow("GoCV Video Capture")
	defer window.Close()
	img := gocv.NewMat()
	defer img.Close()
	yellow := color.RGBA{255, 255, 0, 0}
	fmt.Printf("Start reading camera device\n")
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Cannot read device\n")
			return
		}
		if img.Empty() {
			continue
		}
		imgH := img.Rows()
		imgW := img.Cols()
		gocv.Rectangle(&img, 
			image.Rect(
				imgW/4,
				imgH/4,
				3*imgW/4,
				3*imgH/4,
			),
			yellow,
			2,
		)
		window.IMShow(img)
		if window.WaitKey(1) == 'q' {
			break
		}
	}
}
