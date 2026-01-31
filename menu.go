package main

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type MenuItem struct {
	Label  string
	Action func()
}

func displayMenu(s tcell.Screen, x, y int, title string, items []MenuItem, selected int) {
	styleTitle := tcell.StyleDefault.Foreground(color.Yellow)
	styleNormal := tcell.StyleDefault.Foreground(color.White)
	styleSelected := tcell.StyleDefault.Foreground(color.Black).
		Background(color.Green)

	width, _ := s.Size()
	y = 6

	biggestItem := 0
	for _, item := range items {
		if len(item.Label) > biggestItem {
			biggestItem = len(item.Label)
		}
	}
	boxWidth := biggestItem + 8
	boxHeight := len(items) + 8
	boxX := (width - boxWidth) / 2
	boxY := (y - boxHeight/4)

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
		if i == selected {
			style = styleSelected
		}
		drawTextCentre(s, width, y+i, style, item.Label)
	}
}

func runMenu(s tcell.Screen, title string, items []MenuItem) bool {
	selected := 0

	for {

		w, h := s.Size()
		x := (w - len(title)) / 2
		y := (h - len(items)) / 2

		displayMenu(s, x, y, title, items, selected)
		s.Show()

		ev := <-s.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			s.Clear()
			continue
		case *tcell.EventKey:

			switch ev.Key() {

			case tcell.KeyUp:
				selected--
				if selected < 0 {
					selected = len(items) - 1
				}

			case tcell.KeyDown:
				selected++
				if selected >= len(items) {
					selected = 0
				}

			case tcell.KeyEnter:
				if items[selected].Action != nil {
					items[selected].Action()
				}
				return false
			case tcell.KeyEscape:
				return true
			}
		}
	}
}
