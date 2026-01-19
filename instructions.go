package main

import (
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func displayInstructions(s tcell.Screen, x, y int, title string, items []string) {
	styleTitle := tcell.StyleDefault.Foreground(color.Yellow)
	styleNormal := tcell.StyleDefault.Foreground(color.White)

	width, height := s.Size()
	biggestItem := 0
	for _, item := range items {
		if len(item) > biggestItem {
			biggestItem = len(item)
		}
	}
	boxWidth := biggestItem + 8
	boxHeight := len(items) + 8
	boxX := (width - boxWidth) / 2
	boxY := (height - boxHeight) / 2

	for row := 0; row < boxHeight; row++ {
		for col := 0; col < boxWidth; col++ {
			ch := ' '

			if row == 0 {
				ch = 0x2584
			}
			if row == boxHeight-1 {
				ch = 0x2584
			}
			if col == 0 {
				ch = 0x2590
			}
			if col == boxWidth-1 {
				ch = 0x258C
			}
			if row == 0 && col == 0 {
				ch = 0x2597
			} else if row == 0 && col == boxWidth-1 {
				ch = 0x2596
			} else if row == boxHeight-1 && col == 0 {
				ch = 0x2590
			} else if row == boxHeight-1 && col == boxWidth-1 {
				ch = 0x258c
			}
			s.SetContent(boxX+col, boxY+row, ch, nil, tcell.StyleDefault.Foreground(color.White))
		}
	}
	drawTextCentre(s, width, y, styleTitle, title)
	y += 2

	for i, item := range items {
		style := styleNormal

		drawTextCentre(s, width, y+i, style, item)
	}
}

func runInstructions(s tcell.Screen, title, instructions string) {

	items := strings.Split(instructions, "\n")
	for {

		w, h := s.Size()
		x := (w - len(title)) / 2
		y := (h - len(items)) / 2

		displayInstructions(s, x, y, title, items)
		s.Show()

		ev := <-s.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventKey:

			switch ev.Key() {

			case tcell.KeyEnter, tcell.KeyEscape:
				s.Clear()
				return
			}
		}
	}
}
