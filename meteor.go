package main

import (
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type Meteor struct {
	OldX  float64
	OldY  float64
	X     float64
	Y     float64
	Size  float64
	Speed float64
	Ttl   float64
}

var meteors []Meteor

func updateMeteors() {

	for i := 0; i < len(meteors); i++ {
		meteors[i].OldX = meteors[i].X
		meteors[i].OldY = meteors[i].Y
		meteors[i].Y += meteors[i].Speed
		meteors[i].Ttl -= 1.0
		if meteors[i].Ttl < 0.0 {
			// safe to remove meteor as it will have disappeared
			meteors = append(meteors[:i], meteors[i+1:]...)
			i--
		}
	}
}

func addMeteor(x, y, size, speed, ttl float64) {
	meteor := Meteor{
		X:     x,
		Y:     y,
		Size:  size,
		Speed: speed,
		Ttl:   ttl,
	}
	meteors = append(meteors, meteor)
}
func clearMeteors() {
	meteors = make([]Meteor, 0)
}
func drawMeteors(s tcell.Screen) {
	_, height := s.Size()
	if len(meteors) < 5 || rand.Float64() > 0.99 {
		width, _ := s.Size()
		meteorX := rand.Float64()*float64(width) + 3
		meteorSize := rand.Float64()*2 + 1
		meteorSpeed := rand.Float64()*0.01 + 0.02
		meteorTtl := float64(height) * rand.Float64() * 100
		// meteorX = 20

		addMeteor(meteorX, 0, meteorSize, meteorSpeed, meteorTtl)

	}
	for _, meteor := range meteors {
		for py := 0; py < int(meteor.Size); py++ {
			for px := 0; px < int(meteor.Size); px++ {
				s.SetContent(int(meteor.OldX)+px, int(meteor.OldY)+py, ' ', nil, tcell.StyleDefault.Foreground(color.Black).Background(color.Black))
				if meteor.Ttl > 0.0 && int(meteor.Y)+py < height {
					s.SetContent(int(meteor.X)+px, int(meteor.Y)+py, '*', nil, tcell.StyleDefault.Foreground(color.LightGray).Background(color.Black))
				}

			}
		}
	}
}

func checkForMeteorCollision(shipX, shipY float64) bool {
	shipX = shipX / 2
	shipY = shipY / 2
	for i := range meteors {
		meteor := &meteors[i]
		if shipX+0.75 >= meteor.X && shipX <= meteor.X+meteor.Size {
			if shipY+1 >= meteor.Y && shipY-1 <= meteor.Y+meteor.Size {
				meteor.Ttl = 0.0
				return true
			}
		}
	}
	return false
}
