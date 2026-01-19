package main

import (
	"image"
	"image/draw"
	"log"
	"os"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func getLogoPixels(logo string) []*image.RGBA {
	logoText := make([]rune, 0)
	returnPixels := make([]*image.RGBA, 0)
	for _, char := range logo {
		logoText = append(logoText, char)
	}
	log.Println("logo text", logoText)

	for _, char := range logoText {
		pixels := getPixelsForLetter(char)
		returnPixels = append(returnPixels, pixels)
	}

	return returnPixels
}

func drawLogo(s tcell.Screen, startX int, pixels []*image.RGBA) {

	// logoText := make([]rune, 0)
	// for _, char := range logo {
	// 	logoText = append(logoText, char)
	// }

	startY := 20 //height / 4

	x := startX
	for _, charPixels := range pixels {
		for py := 0; py < charPixels.Bounds().Dy(); py++ {
			for px := 0; px < charPixels.Bounds().Dx(); px++ {
				r, g, b, a := charPixels.At(px, py).RGBA()
				// log.Printf("Pixel at %d,%d : R=%d G=%d B=%d A=%d\n", px, py, r, g, b, a)
				if a > 0 || r > 0 || g > 0 || b > 0 {
					s.SetContent(x+px, startY+py, 'x', nil, tcell.StyleDefault.Foreground(color.Green).Background(color.Black))
					// log.Printf("Drawing pixel at %d,%d : R=%d G=%d B=%d A=%d\n", x+px, startY+py, r, g, b, a)
				}
			}
		}
		x += charPixels.Bounds().Dx() + 2
	}
}

func getPixelsForLetter(character rune) *image.RGBA {
	fontBytes, err := os.ReadFile("Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}
	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
	defer face.Close()

	dr, mask, maskp, _, ok := face.Glyph(fixed.Point26_6{}, character)
	if !ok {
		panic("glyph not found")
	}

	var out *image.RGBA = image.NewRGBA(image.Rect(0, 0, dr.Dx(), dr.Dy()))

	ink := image.NewUniform(color.Black)
	draw.DrawMask(out, out.Bounds(), ink, image.Point{}, mask, maskp, draw.Over)

	// for i := 0; i < len(out.Pix); i += 4 {
	// 	fmt.Printf("Pixel %d: R=%d G=%d B=%d A=%d\n", i/4, out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3])
	// }
	return out
}
