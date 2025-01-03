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

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	fmt.Println("Starting to load page...")

	var buf []byte

	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.nytimes.com/puzzles/letter-boxed"),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.Click("button[data-testid='GDPR-accept']", chromedp.ByQuery, chromedp.AtLeast(0)),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(".lb-board-svg", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
		chromedp.Screenshot(".lb-board-svg", &buf, chromedp.NodeVisible),
	); err != nil {
		log.Printf("Error during page navigation/capture: %v", err)
		return
	}

	fmt.Println("Page loaded and screenshot captured...")

	mat, err := gocv.IMDecode(buf, gocv.IMReadColor)
	if err != nil {
		log.Fatal(err)
	}
	defer mat.Close()

	if ok := gocv.IMWrite("letter_boxed.png", mat); !ok {
		log.Fatal("Failed to save image")
	}

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

	threshold := gocv.NewMat()
	defer threshold.Close()
	gocv.Threshold(gray, &threshold, 127, 255, gocv.ThresholdBinary)

	if ok := gocv.IMWrite("processed.png", threshold); !ok {
		log.Fatal("Failed to save processed image")
	}

	contours := gocv.FindContours(threshold, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	
	green := color.RGBA{0, 255, 0, 255}

	for i := 0; i < contours.Size(); i++ {
		rect := gocv.BoundingRect(contours.At(i))
		if rect.Max.X-rect.Min.X > 10 && rect.Max.Y-rect.Min.Y > 10 { // Filter out tiny contours
			gocv.Rectangle(&mat, rect, green, 2)
		}
	}

	if ok := gocv.IMWrite("detected.png", mat); !ok {
		log.Fatal("Failed to save annotated image")
	}

	fmt.Println("Images have been saved. Check:")
	fmt.Println("1. letter_boxed.png - Original screenshot")
	fmt.Println("2. processed.png - Processed binary image")
	fmt.Println("3. detected.png - Image with detected letter regions marked")
}
