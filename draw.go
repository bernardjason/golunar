package main

import (
	"log"
	"math"

	"github.com/gdamore/tcell/v3"
)

func drawLine(xRunes [][]byte, width, height int, x1, y1, x2, y2 float64, colour byte, orOrxOr bool) {

	if x2 < x1 {
		swap := x1
		x1 = x2
		x2 = swap
	}
	if y2 < y1 {
		swap := y1
		y1 = y2
		y2 = swap
	}
	lineWidth := x2 - x1
	lineHeight := y2 - y1
	stepY := lineHeight / lineWidth
	stepX := 1.0
	if lineHeight > lineWidth {
		stepX = lineWidth / lineHeight
		stepY = 1.0
	}
	if x1 == x2 {
		stepX = 0
	}
	if y1 == y2 {
		stepY = 0
	}

	y := y1
	x := x1
	var bit byte = 0

	oldYi, oldXi := -9999, -9999
	var oldBit byte = 255

	for {
		if x >= x2 && y >= y2 || x <= 0 || x >= float64(width*2) || y <= 0 || y >= float64(height*2) {
			break
		}
		yi := int((y) / 2)
		xi := int((x) / 2)

		Debug("oldBit %x xi %d yi %d\n", oldBit, xi, yi)
		if xRunes[yi] == nil {
			xRunes[yi] = make([]byte, width)
		}

		xx := int(math.Mod(x, 2))
		yy := int(math.Mod(y, 2))

		if xx == 0 && yy == 0 {
			bit = 1
		}
		if xx == 1 && yy == 0 {
			bit = 2
		}
		if xx == 0 && yy == 1 {
			bit = 4
		}
		if xx == 1 && yy == 1 {
			bit = 8
		}
		Debug("b4 rune %x at %f/%f  bit %x\n", xRunes[yi][xi], x, y, bit)
		if bit != oldBit || xi != oldXi || yi != oldYi {
			if orOrxOr {
				xRunes[yi][xi] = xRunes[yi][xi] | bit | colour
			} else {
				xRunes[yi][xi] = xRunes[yi][xi] ^ bit
			}
		}
		Debug("rune %x at %f/%f   bit %x\n", xRunes[yi][xi], x, y, bit)

		oldXi = xi
		oldYi = yi
		oldBit = bit

		y = y + stepY
		x = x + stepX
	}
	// log.Printf("complete\n")
}

func drawRunesToScreen(runes [][]byte, s tcell.Screen, styles [4]tcell.Style) {
	var gl rune = 0
	for yi := range runes {
		for xi := range runes[yi] {
			raw := runes[yi][xi]
			bits := raw & 0x0f
			colour := raw >> 4

			switch bits {
			case 0:
				gl = 0x0020
			case 1:
				gl = 0x2598 // top left
			case 2:
				gl = 0x259d // top right
			case 3:
				gl = 0x2580 // top half
			case 4:
				gl = 0x2596 // lower left
			case 5:
				gl = 0x258c // left half
			case 6:
				gl = 0x259e
			case 7:
				gl = 0x259b
			case 8:
				gl = 0x2597 // lower right
			case 9:
				gl = 0x259a
			case 10:
				gl = 0x2590 // right half
			case 11:
				gl = 0x259c
			case 12:
				gl = 0x2584 // lower half
			case 13:
				gl = 0x2599
			case 14:
				gl = 0x259f
			case 15:
				gl = 0x2588
			default:
				log.Printf("Unexpected %x\n", runes[yi][xi])
			}
			s.SetContent(xi, yi, gl, nil, styles[colour])
		}
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	var width int
	for text != "" {
		text, width = s.Put(col, row, text, style)
		col += width
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
		if width == 0 {
			// incomplete grapheme at end of string
			break
		}
	}
}

func drawShip(xRunes [][]byte, width, height int, xx, yy float64, orOrxOr bool) {

	x := float64(int(xx))
	y := float64(int(yy))

	var colour byte = YELLOW

	drawLine(xRunes, width, height, x-1, y-1, x-1, y+1, colour, orOrxOr)
	drawLine(xRunes, width, height, x+1, y-1, x+1, y+1, colour, orOrxOr)
	drawLine(xRunes, width, height, x, y-1, x+1, y-1, colour, orOrxOr)
	drawLine(xRunes, width, height, x, y-2, x+1, y-2, colour, orOrxOr)

}

func drawTextCentre(s tcell.Screen, width, y int, style tcell.Style, text string) {
	row := y
	col := width/2 - len(text)/2
	x1 := col
	x2 := width - 1

	for text != "" {
		text, width = s.Put(col, row, text, style)
		col += width
		if col >= x2 {
			row++
			col = x1
		}

		if width == 0 {
			// incomplete grapheme at end of string
			break
		}
	}
}
