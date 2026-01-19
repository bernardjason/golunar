package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

const GREEN = 0
const YELLOW = 1 << 4
const RED = 2 << 4
const BLUE = 4 << 4

const instructions = `Welcome to Lunar Lander!

Use the left and right arrow keys to control your lander's thrusters.
Your goal is to land safely on the moon's surface without crashing.

Press Enter or Escape to return to the main menu.`

const title = "LunarLander"

type ExplodeXY struct {
	X, Y       float64
	DirX, DirY float64
	TTL        float64
}
type Explosion struct {
	ExplodeNow int64
	XY         []ExplodeXY
}

var debug = false

func Debug(format string, args ...any) {
	if !debug {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func main() {
	file, err := os.OpenFile(
		"app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println("Problem closing log file", err)
		}
	}()

	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	menu := []MenuItem{

		{
			Label: "Start Game Easy",
			Action: func() {
				runGame(s, 0)
			},
		},
		{
			Label: "Start Game Hard",
			Action: func() {
				runGame(s, 1)
			},
		},
		{
			Label: "Instructions",
			Action: func() {
				runInstructions(s, title, instructions)
			},
		},
		{
			Label: "Quit",
			Action: func() {
				quit()
				os.Exit(0)
			},
		},
	}
	s.Clear()
	var end = false

	width, _ := s.Size()
	startX := width

	click := 0.0
	pixels := getLogoPixels("Bernie soft")
	for !end {
		click++
		if math.Mod(click, 200000) == 0 {
			s.Clear()
			drawLogo(s, startX, pixels)
			s.Show()
			startX -= 1
			if startX < -80 {
				startX = width
			}
			log.Print("Moved logo to ", startX)
		}
		select {
		case ev := <-s.EventQ():
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					end = true
				}
			}
		default:
		}
	}
	s.Clear()
	drawLogo(s, 0, pixels)
	end = false
	for !end {
		end = runMenu(s, title, menu)
	}

}

func runGame(s tcell.Screen, level int) {
	defStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.Reset)

	greenStyle := tcell.StyleDefault.Foreground(color.Green).Background(color.Black)

	thrustStyle := tcell.StyleDefault.Foreground(color.Red).Background(color.Black)

	styles := [...]tcell.Style{
		tcell.StyleDefault.Foreground(color.Green).Background(color.Black),
		tcell.StyleDefault.Foreground(color.Yellow).Background(color.Black),
		tcell.StyleDefault.Foreground(color.Red).Background(color.Black),
		tcell.StyleDefault.Foreground(color.Black).Background(color.Black),
	}

	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	width, height := s.Size()

	var buffer = make([][]byte, height)

	log.Printf("width is %d\n", width)

	landingList := make([]LandingCoOrds, 0)

	if level == 0 {
		landingList = landscapeSin(width, height, buffer, landingList)
	} else {
		landingList = landscapeSinHard(width, height, buffer, landingList)
	}

	log.Printf("Landing points %v\n", landingList)

	playerX, playerY := 10.0, 10.0

	log.Println("Begin game loop")

	drawRunesToScreen(buffer, s, styles)

	var targetGravity = 0.0000017
	var gravity = targetGravity
	var gravityIncrease = targetGravity / float64(height) * 0.0005
	var maxGravity = gravity * float64(height) / 2 * 50
	var speedChangeThrust = maxGravity * 0.07

	var speed = 0.0
	var maxSpeed = maxGravity * 2

	var maximumLandingSpeed = maxSpeed / 4 // 0.00100 // maxSpeed / 8
	const displayMultiplier = 100000.0

	var shouldReturn bool
	var thrust = false
	var displayThrust = 0
	var moveX float64
	var moveY = 0.0
	var landed = false
	var crashed = false
	var setLandedOnce = false

	// easier for debugging without gravity
	var doGravity = true

	var explosion = Explosion{
		ExplodeNow: 0,
		XY:         []ExplodeXY{},
	}
	const ExplosionDone = -200
	const ExplodeBy = 0.05
	explosionDirection := [...]ExplodeXY{
		{X: 0, Y: 0, DirX: ExplodeBy, DirY: ExplodeBy},
		{X: 0, Y: 0, DirX: -ExplodeBy, DirY: ExplodeBy},
		{X: 0, Y: 0, DirX: ExplodeBy, DirY: -ExplodeBy},
		{X: 0, Y: 0, DirX: -ExplodeBy, DirY: -ExplodeBy},
	}
	explosionDirectionIndex := 0

	set := func() {
		if !setLandedOnce {
			setLandedOnce = true
			crashed = true
			explosion.ExplodeNow = 40
			log.Println("")
		}
	}
	setLanded := func() {
		if !setLandedOnce {
			if speed < maximumLandingSpeed {
				setLandedOnce = true
				log.Println("Landed well done")
				landed = true
			} else {
				set()
			}
		}
	}
	onLaunchPad := func(playerX, playerY float64) bool {
		for _, v := range landingList {
			if playerX > float64(v.Start) && playerX < float64(v.End) && math.Abs(playerY-float64(v.Y)) < 2 {
				return true
			}
		}
		return false
	}

loop:

	for explosion.ExplodeNow > ExplosionDone || !crashed && !landed {

		var oddEven = math.Mod(math.Round(float64(int(playerY)+1)), 2)

		drawText(s, 0, 0, 100, 0, greenStyle, fmt.Sprintf("Play lunar lander speed=%.1f   maximum landing speed %.0f   ", speed*displayMultiplier, maximumLandingSpeed*displayMultiplier))

		if !doGravity {
			crashed = false
			landed = false
		}

		select {
		case ev := <-s.EventQ():
			if doGravity {
				thrust, moveX, shouldReturn = handleEventsForGame(ev, s)
			} else {
				moveY, moveX, shouldReturn = handleEventsDebug(ev, s)
			}

			if shouldReturn {
				break loop
			}
		default:
			thrust = false
			moveX = 0.0
			moveY = 0.0
		}
		oldX, oldY := playerX, playerY
		if thrust {
			displayThrust = 200
		}

		if doGravity {
			if !landed && !crashed {
				width, height := s.Size()

				if speed > maxGravity {
					speed = maxGravity
				}
				gravity = gravity + gravityIncrease
				if gravity > maxGravity {
					gravity = maxGravity
				}

				playerY = playerY + speed

				speed = speed + gravity

				playerX = playerX + moveX
				if playerX < 0 || int(playerX) >= width*2 || playerY < 0 || int(playerY) >= height*2 {
					playerX, playerY = oldX, oldY
				}

				if thrust {
					speed = speed - speedChangeThrust
					gravity = 0 //gravity * 0.5
					if speed < -maxSpeed {
						speed = -maxSpeed
					}

				}

			}
		} else {
			playerX = playerX + moveX
			playerY = playerY + moveY
		}

		if playerY >= float64(height*2)-2 || playerY <= 2 {
			set()
		}

		// erase here so thrust up doesn't trip collision detection
		if !crashed {
			drawShip(buffer, width, height, float64(oldX), float64(oldY), false)
		}

		checkCollisionBelow(oddEven, playerY, buffer, playerX, onLaunchPad, setLanded, set)
		checkCollisionAbove(oddEven, playerY, buffer, playerX, set)

		if !crashed {
			drawShip(buffer, width, height, float64(playerX), float64(playerY), true)
		}

		drawRunesToScreen(buffer, s, styles)

		if explosion.ExplodeNow > 0 {
			log.Println("Explosion step", explosion.ExplodeNow)
			explosionDirectionIndex = int(math.Mod(float64(explosionDirectionIndex+1), float64(len(explosionDirection))))

			explosion.XY = append(explosion.XY, ExplodeXY{
				X:    playerX,
				Y:    playerY,
				DirX: explosionDirection[explosionDirectionIndex].DirX * (math.Mod(rand.Float64(), 1.0)),
				DirY: explosionDirection[explosionDirectionIndex].DirY * (math.Mod(rand.Float64(), 1.0)),
				TTL:  math.Mod(rand.Float64()*1000, 500) + 100,
			})
			log.Println("Added explosion part", len(explosion.XY))
		}

		for i := range explosion.XY {
			update := &explosion.XY[i]
			bufferIndexX := int(update.X / 2)
			bufferIndexY := int(update.Y / 2)
			if update.TTL > 0 && update.Y > 0 && bufferIndexY < len(buffer) && bufferIndexX >= 0 && bufferIndexX < width {
				if buffer[bufferIndexY] == nil {
					buffer[bufferIndexY] = make([]byte, width)
				}

				buffer[bufferIndexY][bufferIndexX] = 0
				update.TTL--
				Debug("Update explosion", i, update.X, update.Y, update.DirX, update.DirY)
				update.X = update.X + update.DirX
				update.Y = update.Y + update.DirY
				bufferIndexX := int(update.X / 2)
				bufferIndexY := int(update.Y / 2)
				if bufferIndexY < len(buffer) && bufferIndexX >= 0 && bufferIndexX < width {
					if buffer[bufferIndexY] == nil {
						buffer[bufferIndexY] = make([]byte, width)
					}
					buffer[bufferIndexY][bufferIndexX] = 15 | RED
					Debug("Added buffer entry", bufferIndexY, update.X, buffer[bufferIndexY][bufferIndexX])
				}
			}
		}
		explosion.ExplodeNow--
		if explosion.ExplodeNow <= ExplosionDone {
			for i := range explosion.XY {
				update := &explosion.XY[i]
				if update.Y >= 0 && update.Y < float64(height) {
					buffer[int(update.Y/2)][int(update.X/2)] = 0
				}
			}
			explosion.XY = []ExplodeXY{}
		}

		if displayThrust > 0 || !doGravity {

			displayThrust--

			removePreviousThrust(oldY, oldX, s, thrustStyle)
			if displayThrust > 1 || !doGravity {
				drawThrust(playerX, playerY, displayThrust, s, thrustStyle)
			}
		}

		if landed {
			if speed < maximumLandingSpeed {
				drawText(s, 10, 10, 200, 10, greenStyle, fmt.Sprintf("Well done. Speed was %.1f   target=%.1f  maximum speed %0.1f           ", speed*displayMultiplier, maximumLandingSpeed*displayMultiplier, maxGravity*displayMultiplier))
			} else {
				set()
			}
		}
		if crashed {
			drawText(s, 10, 10, 200, 10, greenStyle, fmt.Sprintf("Crashed. Speed was %.1f   target speed %.1f  maximum speed %.1f                ", speed*displayMultiplier, maximumLandingSpeed*displayMultiplier, maxGravity*displayMultiplier))
		}

		s.Show()
	}
}

func drawThrust(playerX float64, playerY float64, displayThrust int, s tcell.Screen, thrustStyle tcell.Style) {
	oddEvenX := math.Mod(math.Round(float64(int(playerX))), 2)
	thrustY := int(playerY)/2 + 1
	thrustX := int(playerX-1) / 2
	const fullBlock1 = 0x2591
	const fullBlock2 = 0x259a
	const fullBlock3 = 0x259e
	const halfLeft1 = 0x258C
	const halfLeft2 = 0x2598
	const halfLeft3 = 0x2596
	const halfRight1 = 0x259d
	const halfRight2 = 0x2590
	const halfRight3 = 0x2597
	thrustCharactersFullBlockHalfLeft := []rune{fullBlock1, halfLeft1, fullBlock2, halfLeft2, fullBlock3, halfLeft3}
	thrustCharactersHalfRightFullBlock := []rune{halfRight1, fullBlock1, halfRight2, fullBlock2, halfRight3, fullBlock3}
	thrustCharacter := int(math.Mod(float64(displayThrust), float64(len(thrustCharactersFullBlockHalfLeft)/2))) * 2
	if thrustCharacter < 0 {
		thrustCharacter = 0
	}

	if oddEvenX == 1 {
		s.SetContent(thrustX, thrustY, thrustCharactersFullBlockHalfLeft[thrustCharacter], nil, thrustStyle)
		s.SetContent(thrustX+1, thrustY, thrustCharactersFullBlockHalfLeft[thrustCharacter+1], nil, thrustStyle)
	} else {
		s.SetContent(thrustX, thrustY, thrustCharactersHalfRightFullBlock[thrustCharacter], nil, thrustStyle)
		s.SetContent(thrustX+1, thrustY, thrustCharactersHalfRightFullBlock[thrustCharacter+1], nil, thrustStyle)
	}
}

func removePreviousThrust(oldY float64, oldX float64, s tcell.Screen, thrustStyle tcell.Style) {
	thrustY := int(oldY)/2 + 1
	thrustX := int(oldX-1) / 2
	s.SetContent(thrustX, thrustY, ' ', nil, thrustStyle)
	s.SetContent(thrustX+1, thrustY, ' ', nil, thrustStyle)
}

func checkCollisionAbove(oddEven float64, playerY float64, buffer [][]byte, playerX float64, set func()) {
	var yi int
	if oddEven == 0 {
		yi = int(int(playerY-2) / 2)
	} else {
		yi = int(int(playerY-3) / 2)
	}

	var singleByte byte
	if buffer[yi] != nil && buffer[yi][int(playerX)/2] != 0 {
		if oddEven == 0 {
			singleByte = buffer[yi][int(playerX/2)] & 3
			if singleByte != 0 {
				log.Println("!!!!!!!!!!! 1", buffer[yi][int(playerX/2)])
				set()
			}
		} else {
			singleByte = buffer[yi][int(playerX/2)] & 12
			if singleByte != 0 {
				log.Println("!!!!!!!!!!! 2", buffer[yi][int(playerX/2)])
				set()
			}
		}
	}
}

func checkCollisionBelow(oddEven float64, playerY float64, buffer [][]byte, playerX float64, onLaunchPad func(playerX float64, playerY float64) bool, setLanded func(), set func()) {
	var yi int
	if oddEven == 1 {
		yi = int(int(playerY) / 2)
	} else {
		yi = int(int(playerY+1) / 2)
	}

	var singleByte byte
	if buffer[yi] != nil && buffer[yi][int(playerX)/2] != 0 {
		if oddEven == 1 {
			singleByte = buffer[yi][int(playerX/2)] & 12
			if singleByte != 0 {
				if singleByte == 12 && onLaunchPad(playerX, playerY) {
					setLanded()
				} else {
					set()
				}
			}
		} else {
			singleByte = buffer[yi][int(playerX/2)] & 3
			if singleByte != 0 {
				if singleByte == 3 && onLaunchPad(playerX, playerY) {
					setLanded()
				} else {
					set()
				}
			}
		}
	}
}

func handleEventsForGame(ev tcell.Event, s tcell.Screen) (bool, float64, bool) {
	var playerX, thrust = 0.0, false

	switch ev := ev.(type) {
	case *tcell.EventResize:
		s.Sync()
	case *tcell.EventKey:

		if ev.Modifiers() == 1 {
			thrust = true
		}

		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
			return false, 0.0, true
		} else if ev.Key() == tcell.KeyCtrlL {
			s.Sync()
		}

		key := ev.Key()
		switch key {
		case tcell.KeyUp:
			thrust = true
		case tcell.KeyLeft:
			playerX = -1
		case tcell.KeyRight:
			playerX = +1
		}

	}
	return thrust, playerX, false
}

func handleEventsDebug(ev tcell.Event, s tcell.Screen) (float64, float64, bool) {
	var playerX, playerY = 0.0, 0.0
	switch ev := ev.(type) {
	case *tcell.EventResize:
		s.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
			return 0.0, 0.0, true
		} else if ev.Key() == tcell.KeyCtrlL {
			s.Sync()
		}
		key := ev.Key()
		switch key {
		case tcell.KeyUp:
			playerY = -1
		case tcell.KeyDown:
			playerY = 1
		case tcell.KeyLeft:
			playerX = -1
		case tcell.KeyRight:
			playerX = +1
		}

	}
	return playerY, playerX, false
}
