package main

import (
	"log"
	"math"
)

type LandingCoOrds struct {
	Start  int
	End    int
	Y      int
	Points int
}

func landscapeSin(width int, height int, buffer [][]byte, landingList []LandingCoOrds) []LandingCoOrds {
	var currentLandingEntry = LandingCoOrds{Start: 0, End: -1, Y: -1, Points: 0}
	angle := 0.0
	var oldX = 0
	var oldY = 0.0
	for x := 0; x < width*2; x = x + 1 {
		y := float64(height) + (math.Sin(angle) * float64(height/3)) + float64(height/2)
		angle = angle + 0.025
		drawLine(buffer, width, height, float64(oldX), float64(oldY), float64(x), float64(y), GREEN, true)
		oldX = x
		oldY = y
		if int(y) != currentLandingEntry.Y {
			if currentLandingEntry.Points > 6 {
				currentLandingEntry.End = x - 1
				landingList = append(landingList, currentLandingEntry)
				showLandingSite(currentLandingEntry, buffer, width, height)
			}
			currentLandingEntry = LandingCoOrds{Start: x, End: -1, Y: int(y), Points: 0}
		} else {
			currentLandingEntry.Points++
		}
	}
	return landingList
}

func landscapeSinHard(width int, height int, buffer [][]byte, landingList []LandingCoOrds) []LandingCoOrds {
	var currentLandingEntry = LandingCoOrds{Start: 0, End: -1, Y: -1, Points: 0}
	angle := 0.0
	var oldX = 0
	var oldY = 0.0
	for x := 0; x < width*2; x = x + 1 {
		y := float64(height) + (math.Sin(angle) * float64(height/3)) + float64(height/2)
		angle = angle + 0.025
		drawLine(buffer, width, height, float64(oldX), float64(oldY), float64(x), float64(y), GREEN, true)
		oldX = x
		oldY = y
		if int(y) != currentLandingEntry.Y && float64(x) > float64(width)*1.5 {
			if currentLandingEntry.Points > 6 {
				currentLandingEntry.End = x - 1
				landingList = append(landingList, currentLandingEntry)
				showLandingSite(currentLandingEntry, buffer, width, height)
			}
			currentLandingEntry = LandingCoOrds{Start: x, End: -1, Y: int(y), Points: 0}
		} else {
			currentLandingEntry.Points++
		}
	}

	oldY = float64(height) + (math.Sin(angle) * float64(height/3)) + float64(height/2) - 20
	oldX = 20
	angle = 0
	for x := 20; x < width*2; x = x + 1 {
		y := float64(height) + (math.Sin(angle) * float64(height/3)) + float64(height/2) - 30
		angle = angle + 0.025
		drawLine(buffer, width, height, float64(oldX), float64(oldY), float64(x), float64(y), GREEN, true)
		oldX = x
		oldY = y
	}
	return landingList
}

func showLandingSite(currentLandingEntry LandingCoOrds, buffer [][]byte, width int, height int) {
	startYHere := currentLandingEntry.Y + 1
	colourList := [...]byte{RED, GREEN, GREEN, RED}
	colourListIndex := 0
	if math.Mod(float64(currentLandingEntry.Y), 2) == 0 {
		colourListIndex++
	}
	for y := startYHere; y < startYHere+5; y++ {
		drawLine(buffer, width, height, float64(currentLandingEntry.Start), float64(y), float64(currentLandingEntry.End), float64(y), colourList[colourListIndex], true)
		colourListIndex++
		if colourListIndex >= len(colourList) {
			colourListIndex = 0
		}
	}
}

func landscapeHard(width int, height int, buffer [][]byte, landingList []LandingCoOrds) []LandingCoOrds {
	type XY struct {
		StartX float64
		StartY float64
		EndX   float64
		EndY   float64
	}
	h := float64(height * 2)
	w := float64(width * 2)
	landingPadX := 50.0
	landingPadY := h * 0.70
	coords := [...]XY{

		{21, h * .25, 21, h},
		{21, h * 0.55, w * 0.5, h * 0.55},
		{w * 0.65, h * 0.55, w, h * 0.55},
		{w * 0.5, h * 0.55, w * 0.5, h * 0.55},
		{w * 0.5, h * 0.80, 40, h * 0.80},
		{1, h * 0.75, 10, h * 0.75},
		{w * 0.5, h * 0.55, w * 0.5, h * 0.80},
		{landingPadX, landingPadY, landingPadX + 20, landingPadY},
	}
	var currentLandingEntry = LandingCoOrds{Start: int(landingPadX), End: int(landingPadX) + 20, Y: int(landingPadY), Points: 10}

	landingList = append(landingList, currentLandingEntry)
	showLandingSite(currentLandingEntry, buffer, width, height)

	for _, v := range coords {
		drawLine(buffer, width, height, v.StartX, v.StartY, v.EndX, v.EndY, GREEN, true)
		log.Printf("draw line %f,%f to %f,%f\n", v.StartX, v.StartY, v.EndX, v.EndY)
	}

	return landingList
}
