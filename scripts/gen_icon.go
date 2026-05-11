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

	// Background: Deep Charcoal
	bgCol := color.RGBA{18, 18, 18, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgCol}, image.Point{}, draw.Src)

	// Colors
	neonBlue := color.RGBA{0, 210, 255, 255}
	neonCyan := color.RGBA{0, 255, 255, 255}
	windowGray := color.RGBA{45, 45, 45, 255}
	dotRed := color.RGBA{255, 95, 87, 255}
	dotYellow := color.RGBA{255, 189, 46, 255}
	dotGreen := color.RGBA{39, 201, 63, 255}

	// 1. Draw Stylized Terminal Window
	// Main body
	drawRoundedRect(img, 30, 40, 226, 216, 12, windowGray)
	// Top bar
	drawRoundedRect(img, 30, 40, 226, 75, 12, color.RGBA{60, 60, 60, 255})
	// Cover bottom rounding of top bar
	drawRect(img, 30, 65, 226, 75, color.RGBA{60, 60, 60, 255})

	// 2. Draw Window Controls (dots)
	drawCircleFilled(img, 50, 57, 6, dotRed)
	drawCircleFilled(img, 70, 57, 6, dotYellow)
	drawCircleFilled(img, 90, 57, 6, dotGreen)

	// 3. Draw Prompt '>'
	// Using the neon cyan for the prompt
	drawPrompt(img, 70, 110, 40, neonCyan)

	// 4. Draw Refresh Arrow wrapping around the prompt
	// Circular path from ~200 degrees to ~160 degrees
	drawRefreshArrow(img, 128, 145, 60, 12, neonBlue)

	// 5. Save Icon
	f, _ := os.Create("winres/icon.png")
	defer f.Close()
	png.Encode(f, img)

	// 6. Create 16x16 version
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

func drawRoundedRect(img *image.RGBA, x1, y1, x2, y2, r int, col color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			// Check if pixel is in rounded corners
			inCorner := false
			if x < x1+r && y < y1+r { // Top-left
				if math.Sqrt(math.Pow(float64(x-(x1+r)), 2)+math.Pow(float64(y-(y1+r)), 2)) > float64(r) {
					inCorner = true
				}
			} else if x > x2-r-1 && y < y1+r { // Top-right
				if math.Sqrt(math.Pow(float64(x-(x2-r-1)), 2)+math.Pow(float64(y-(y1+r)), 2)) > float64(r) {
					inCorner = true
				}
			} else if x < x1+r && y > y2-r-1 { // Bottom-left
				if math.Sqrt(math.Pow(float64(x-(x1+r)), 2)+math.Pow(float64(y-(y2-r-1)), 2)) > float64(r) {
					inCorner = true
				}
			} else if x > x2-r-1 && y > y2-r-1 { // Bottom-right
				if math.Sqrt(math.Pow(float64(x-(x2-r-1)), 2)+math.Pow(float64(y-(y2-r-1)), 2)) > float64(r) {
					inCorner = true
				}
			}

			if !inCorner {
				img.Set(x, y, col)
			}
		}
	}
}

func drawCircleFilled(img *image.RGBA, x0, y0, r int, col color.Color) {
	for x := x0 - r; x <= x0+r; x++ {
		for y := y0 - r; y <= y0+r; y++ {
			if math.Sqrt(math.Pow(float64(x-x0), 2)+math.Pow(float64(y-y0), 2)) <= float64(r) {
				img.Set(x, y, col)
			}
		}
	}
}

func drawPrompt(img *image.RGBA, x, y, size int, col color.Color) {
	thickness := size / 4
	for i := 0; i < size; i++ {
		for t := 0; t < thickness; t++ {
			// Upper diagonal
			img.Set(x+i, y+i+t, col)
			// Lower diagonal
			img.Set(x+i, y+size*2-i+t, col)
		}
	}
}

func drawRefreshArrow(img *image.RGBA, x0, y0, r, width int, col color.Color) {
	// Draw the arc
	for x := x0 - r - width; x <= x0+r+width; x++ {
		for y := y0 - r - width; y <= y0+r+width; y++ {
			dist := math.Sqrt(float64((x-x0)*(x-x0) + (y-y0)*(y-y0)))
			if dist >= float64(r-width/2) && dist <= float64(r+width/2) {
				angle := math.Atan2(float64(y-y0), float64(x-x0))
				// Nearly complete circle, leaving a gap for the arrow head
				if angle < 0.5 || angle > 1.2 {
					img.Set(x, y, col)
				}
			}
		}
	}

	// Draw the arrow head
	// Positioned at the end of the arc
	headX, headY := x0+int(float64(r)*math.Cos(0.5)), y0+int(float64(r)*math.Sin(0.5))
	limit := int(float64(width) * 1.5)
	for dx := -width; dx <= width; dx++ {
		for dy := -width; dy <= width; dy++ {
			// A simple triangle-ish head
			if int(math.Abs(float64(dx))+math.Abs(float64(dy))) <= limit {
				img.Set(headX+dx, headY+dy, col)
			}
		}
	}
}
