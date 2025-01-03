package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"gocv.io/x/gocv"
)

func main() {
	// Create a new chrome instance with options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Create a timeout - increased to 45 seconds
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	fmt.Println("Starting to load page...")

	// Buffer to store the screenshot
	var buf []byte

	// Navigate to NYT Letter Boxed with more robust loading sequence
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.nytimes.com/puzzles/letter-boxed"),
		// Wait for initial page load
		chromedp.WaitVisible("body"),
		// Add a small delay
		chromedp.Sleep(2*time.Second),
		// Try to handle possible cookie dialog
		chromedp.Click("button[data-testid='GDPR-accept']", chromedp.ByQuery, chromedp.AtLeast(0)),
		// Add another small delay
		chromedp.Sleep(2*time.Second),
		// Wait for game board
		chromedp.WaitVisible(".lb-board-svg", chromedp.ByQuery),
		// Final delay to ensure full render
		chromedp.Sleep(3*time.Second),
		// Capture screenshot
		chromedp.Screenshot(".lb-board-svg", &buf, chromedp.NodeVisible),
	); err != nil {
		log.Printf("Error during page navigation/capture: %v", err)
		return
	}

	fmt.Println("Page loaded and screenshot captured...")

	// Decode the screenshot buffer into a Mat
	mat, err := gocv.IMDecode(buf, gocv.IMReadColor)
	if err != nil {
		log.Fatal(err)
	}
	defer mat.Close()

	// Save raw screenshot
	if ok := gocv.IMWrite("letter_boxed.png", mat); !ok {
		log.Fatal("Failed to save image")
	}

	// Convert to grayscale
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

	// Apply thresholding to make text more visible
	threshold := gocv.NewMat()
	defer threshold.Close()
	gocv.Threshold(gray, &threshold, 127, 255, gocv.ThresholdBinary)

	// Save processed image
	if ok := gocv.IMWrite("processed.png", threshold); !ok {
		log.Fatal("Failed to save processed image")
	}

	// Find contours (letter shapes)
	contours := gocv.FindContours(threshold, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	
	// Define color for rectangles
	green := color.RGBA{0, 255, 0, 255}

	// Draw rectangles around detected letters
	for i := 0; i < contours.Size(); i++ {
		rect := gocv.BoundingRect(contours.At(i))
		if rect.Max.X-rect.Min.X > 10 && rect.Max.Y-rect.Min.Y > 10 { // Filter out tiny contours
			gocv.Rectangle(&mat, rect, green, 2)
		}
	}

	// Save annotated image
	if ok := gocv.IMWrite("detected.png", mat); !ok {
		log.Fatal("Failed to save annotated image")
	}

	fmt.Println("Images have been saved. Check:")
	fmt.Println("1. letter_boxed.png - Original screenshot")
	fmt.Println("2. processed.png - Processed binary image")
	fmt.Println("3. detected.png - Image with detected letter regions marked")
}
