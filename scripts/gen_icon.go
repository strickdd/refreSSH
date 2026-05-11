package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	xdraw "golang.org/x/image/draw"
)

func main() {
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background: Rounded dark square
	bgCol := color.RGBA{30, 30, 30, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgCol}, image.Point{}, draw.Src)

	// Primary color from config: #43ED0F
	accentCol := color.RGBA{0x43, 0xED, 0x0F, 255}

	// Draw a simple ">" prompt
	// Horizontal part
	drawRect(img, 60, 100, 120, 115, accentCol)
	// Diagonal down
	for i := 0; i < 60; i++ {
		drawRect(img, 60+i, 100+i, 60+i+15, 100+i+15, accentCol)
	}
	// Diagonal up
	for i := 0; i < 60; i++ {
		drawRect(img, 120-i, 160+i, 120-i+15, 160+i+15, accentCol)
	}

	// Draw a refresh arrow (circle with gap)
	drawCircle(img, 180, 180, 40, 10, accentCol)
	// Arrow head
	drawRect(img, 180+30, 180-10, 180+50, 180+10, accentCol)

	f, _ := os.Create("winres/icon.png")
	defer f.Close()
	png.Encode(f, img)

	// Also create a 16x16 version
	img16 := image.NewRGBA(image.Rect(0, 0, 16, 16))
	xdraw.BiLinear.Scale(img16, img16.Bounds(), img, img.Bounds(), draw.Over, nil)
	f16, _ := os.Create("winres/icon16.png")
	defer f16.Close()
	png.Encode(f16, img16)
}

func drawRect(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, col)
		}
	}
}

func drawCircle(img *image.RGBA, x0, y0, r, width int, col color.Color) {
	for x := x0 - r - width; x <= x0+r+width; x++ {
		for y := y0 - r - width; y <= y0+r+width; y++ {
			dist := math.Sqrt(float64((x-x0)*(x-x0) + (y-y0)*(y-y0)))
			if dist >= float64(r) && dist <= float64(r+width) {
				// Create a gap for the arrow look
				angle := math.Atan2(float64(y-y0), float64(x-x0))
				if angle < -0.5 || angle > 0.5 {
					img.Set(x, y, col)
				}
			}
		}
	}
}
