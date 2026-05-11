package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background: Pure Black
	bgCol := color.RGBA{0, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgCol}, image.Point{}, draw.Src)

	// Primary color: Bright Terminal Green (#43ED0F)
	green := color.RGBA{0x43, 0xED, 0x0F, 255}

	// Draw text "R>_"
	s := "R>_"
	
	// Create a temporary canvas for the font
	textImg := image.NewRGBA(image.Rect(0, 0, 30, 15))
	d := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(green),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(2), Y: fixed.I(11)},
	}
	d.DrawString(s)

	// Scale up to fill the 256x256 icon
	// We want to preserve sharp pixels, so use NearestNeighbor
	xdraw.NearestNeighbor.Scale(img, img.Bounds(), textImg, textImg.Bounds(), draw.Over, nil)

	// Save Icon
	f, _ := os.Create("winres/icon.png")
	defer f.Close()
	png.Encode(f, img)

	// Create 16x16 version
	img16 := image.NewRGBA(image.Rect(0, 0, 16, 16))
	xdraw.NearestNeighbor.Scale(img16, img16.Bounds(), textImg, textImg.Bounds(), draw.Over, nil)
	f16, _ := os.Create("winres/icon16.png")
	defer f16.Close()
	png.Encode(f16, img16)
}
