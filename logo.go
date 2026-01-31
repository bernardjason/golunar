package main

import (
	"embed"
	"image"
	"image/draw"
	"log"

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

func drawLogo(s tcell.Screen, startX int, pixels []*image.RGBA) bool {

	_, startY := s.Size()
	startY = startY / 10

	letterColourList := []tcell.Color{color.Red, color.Green, color.Blue, color.Yellow, color.LightCyan, color.DarkMagenta, color.White}
	letterColourIndex := 0
	x := startX
	for _, charPixels := range pixels {
		for py := 0; py < charPixels.Bounds().Dy(); py++ {
			for px := 0; px < charPixels.Bounds().Dx(); px++ {
				r, g, b, a := charPixels.At(px, py).RGBA()
				if a > 0 || r > 0 || g > 0 || b > 0 {
					s.SetContent(x+px, startY+py, 'x', nil, tcell.StyleDefault.Foreground(letterColourList[letterColourIndex]).Background(color.Black))
				}
			}
		}
		letterColourIndex++
		if letterColourIndex >= len(letterColourList) {
			letterColourIndex = 0
		}
		x += charPixels.Bounds().Dx() + 2
	}
	if x < 0 {
		return true
	}
	return false
}

//go:embed Roboto-Regular.ttf
var fontFile embed.FS

func getPixelsForLetter(character rune) *image.RGBA {
	fontBytes, err := fontFile.ReadFile("Roboto-Regular.ttf")
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

	dr, mask, maskP, _, ok := face.Glyph(fixed.Point26_6{}, character)
	if !ok {
		panic("glyph not found")
	}

	var out *image.RGBA = image.NewRGBA(image.Rect(0, 0, dr.Dx(), dr.Dy()))

	ink := image.NewUniform(color.Black)
	draw.DrawMask(out, out.Bounds(), ink, image.Point{}, mask, maskP, draw.Over)

	return out
}
