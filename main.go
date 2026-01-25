package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

const GREEN = 0
const YELLOW = 1 << 4
const RED = 2 << 4
const BLUE = 4 << 4

const instructions = `Welcome to Lunar Lander!

Use the left and right arrow keys to control your lander's.
Up arrow fires thruster. Press shift with left right also fires thrusters.
Avoid the meteors!
Your goal is to land safely on the moon's surface without crashing and as
much fuel as possible.

Press Enter or Escape to return to the main menu.`

const title = "LunarLander"

type ExplodeXY struct {
	X, Y       float64
	DirX, DirY float64
	TTL        float64
}
type Explosion struct {
	MeteorHit  bool
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
	if !IsWASM {
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
	}
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
	if IsWASM {
		for i := range menu {
			if menu[i].Label == "Quit" {
				menu = append(menu[:i], menu[i+1:]...)
				break
			}
		}
	}
	s.Clear()
	var end = false

	width, _ := s.Size()
	startX := width

	click := 0.0
	var targetFps int64 = 30

	startTime := time.Now()

	pixels := getLogoPixels("BERNIESOFT")

	for !end {

		click++
		s.Clear()
		drawLogo(s, startX, pixels)
		s.Show()
		startX -= 1
		if startX < -200 {
			startX = width
			log.Print("Moved logo to ", startX)
		}

		select {
		case ev := <-s.EventQ():
			switch ev.(type) {
			case *tcell.EventKey:
				end = true
			}
		default:
		}
		checkAverageFps(startTime, click, targetFps)
	}
	s.Clear()
	drawLogo(s, startX, pixels)

	end = false
	for !end {
		end = runMenu(s, title, menu)
		// is WASM don't allow quit from menu
		if IsWASM {
			end = false
		}
		checkAverageFps(startTime, click, targetFps)
	}

}

func checkAverageFps(startTime time.Time, click float64, targetFps int64) {
	if click > 0 {
		now := time.Now()
		averageFrameTimeInMicroSeconds := now.Sub(startTime).Microseconds() / int64(click)

		Debug("Average frame time %.2f microseconds over %d clicks\n", float64(averageFrameTimeInMicroSeconds), int(click))
		if averageFrameTimeInMicroSeconds < time.Second.Microseconds()/(targetFps) {
			pause := time.Duration((time.Second.Microseconds()/(targetFps) - averageFrameTimeInMicroSeconds)) * time.Microsecond
			time.Sleep(pause)
		}
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
		// landingList = landscapeSinHard(width, height, buffer, landingList)
		landingList = landscapeHard(width, height, buffer, landingList)
	}

	log.Printf("Landing points %v\n", landingList)

	playerX, playerY := 10.0, 10.0

	log.Println("Begin game loop")

	drawRunesToScreen(buffer, s, styles)

	var targetGravity = 0.000017
	var gravity = targetGravity
	var gravityIncrease = targetGravity / float64(height) * 0.005
	var maxGravity = gravity * float64(height) / 2 * 50
	var speedChangeThrust = maxGravity * 0.01
	var fuel = 100.0
	var hits = 0
	const permittedHits = 3

	var speed = 0.0
	var maxSpeed = maxGravity

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
	var ExplosionDone int64 = -200
	const ExplodeBy = 0.05
	explosionDirection := [...]ExplodeXY{
		{X: 0, Y: 0, DirX: ExplodeBy, DirY: ExplodeBy},
		{X: 0, Y: 0, DirX: -ExplodeBy, DirY: ExplodeBy},
		{X: 0, Y: 0, DirX: ExplodeBy, DirY: -ExplodeBy},
		{X: 0, Y: 0, DirX: -ExplodeBy, DirY: -ExplodeBy},
	}
	explosionDirectionIndex := 0

	setCrashed := func() {
		if !setLandedOnce {
			setLandedOnce = true
			crashed = true
			explosion.ExplodeNow = 40
			explosion.MeteorHit = false
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
				setCrashed()
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

	clearMeteors()
	startTime := time.Now()
	var targetFps int64 = 60
	click := 0.0
loop:

	for explosion.ExplodeNow > ExplosionDone || !crashed && !landed {

		click++
		var oddEven = math.Mod(math.Round(float64(int(playerY)+1)), 2)

		drawText(s, 0, 0, 100, 0, greenStyle, fmt.Sprintf("Play lunar lander speed=%.1f maximum landing speed %.0f fuel %0.f hits %d   ", speed*displayMultiplier, maximumLandingSpeed*displayMultiplier, fuel, hits))

		updateMeteors()

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

				if fuel > 0 && thrust {
					fuel = fuel - 0.5
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
			setCrashed()
		}

		// erase here so thrust up doesn't trip collision detection
		if !crashed {
			drawShip(buffer, width, height, float64(oldX), float64(oldY), false)
		}

		checkCollisionBelow(oddEven, playerY, buffer, playerX, onLaunchPad, setLanded, setCrashed)
		checkCollisionAbove(oddEven, playerY, buffer, playerX, setCrashed)
		if checkForMeteorCollision(playerX, playerY) {
			explosion.ExplodeNow = 10
			explosion.MeteorHit = true
			hits++
			if hits >= permittedHits {
				setCrashed()
			}
		}

		if !crashed {
			drawShip(buffer, width, height, float64(playerX), float64(playerY), true)
		}

		drawRunesToScreen(buffer, s, styles)

		drawMeteors(s)

		// handleExplosion(explosion, buffer, width, ExplosionDone, height)
		if explosion.ExplodeNow > 0 {
			log.Println("Explosion step", explosion.ExplodeNow)
			explosionDirectionIndex = int(math.Mod(float64(explosionDirectionIndex+1), float64(len(explosionDirection))))

			toAdd := ExplodeXY{
				X:    playerX / 2,
				Y:    playerY / 2,
				DirX: explosionDirection[explosionDirectionIndex].DirX * (math.Mod(rand.Float64(), 1.0)),
				DirY: explosionDirection[explosionDirectionIndex].DirY * (math.Mod(rand.Float64(), 1.0)),
				TTL:  math.Mod(rand.Float64()*1000, 500) + 100,
			}
			if explosion.MeteorHit {
				toAdd.X = toAdd.X + toAdd.DirX/toAdd.DirX*2
				toAdd.Y = toAdd.Y + toAdd.DirY/toAdd.DirY*2
			}

			explosion.XY = append(explosion.XY, toAdd)
			log.Println("Added explosion part", len(explosion.XY))
		}

		for i := range explosion.XY {
			update := &explosion.XY[i]
			s.SetContent(int(update.X), int(update.Y), ' ', nil, tcell.StyleDefault.Foreground(color.Red).Background(color.Black))
			update.X = update.X + update.DirX
			update.Y = update.Y + update.DirY
			update.TTL--
			if update.TTL <= 0 {
				update.X = -1
				update.Y = -1
			}
			s.SetContent(int(update.X), int(update.Y), '*', nil, tcell.StyleDefault.Foreground(color.Red).Background(color.Black))
		}

		if explosion.ExplodeNow <= ExplosionDone {
			for i := range explosion.XY {
				update := &explosion.XY[i]
				s.SetContent(int(update.X), int(update.Y), ' ', nil, tcell.StyleDefault.Foreground(color.Red).Background(color.Black))
			}
			explosion.XY = []ExplodeXY{}
		}
		explosion.ExplodeNow--

		if displayThrust > 0 || !doGravity {

			displayThrust--

			removePreviousThrust(oldY, oldX, s, thrustStyle)
			if displayThrust > 1 || !doGravity {
				drawThrust(playerX, playerY, displayThrust, s, thrustStyle)
			}
		}

		if landed {
			if speed < maximumLandingSpeed {
				score := (fuel + 1) / (speed + 1) / float64(hits+1)
				drawText(s, 5, 10, 200, 10, greenStyle, fmt.Sprintf("Well done. Score %0.f speed was %.1f fuel %0.f hits %d ", score, speed*displayMultiplier, fuel, hits))
			} else {
				setCrashed()
			}
		}
		if crashed {
			drawText(s, 5, 10, 200, 10, greenStyle, fmt.Sprintf("Crashed. Speed was %.1f target speed %.1f fuel %0.f hits %d", speed*displayMultiplier, maximumLandingSpeed*displayMultiplier, fuel, hits))
		}

		s.Show()
		checkAverageFps(startTime, click, targetFps)
	}
}

func handleExplosion(explosion Explosion, buffer [][]byte, width int, ExplosionDone int64, height int) {
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

	if explosion.ExplodeNow <= ExplosionDone {
		for i := range explosion.XY {
			update := &explosion.XY[i]
			if update.Y >= 0 && update.Y < float64(height) && int(update.X/2) >= 0 && int(update.X/2) < width {
				buffer[int(update.Y/2)][int(update.X/2)] = 0
			}
		}
		explosion.XY = []ExplodeXY{}
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

func checkCollisionAbove(oddEven float64, playerY float64, buffer [][]byte, playerX float64, setCrashed func()) {
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
				setCrashed()
			}
		} else {
			singleByte = buffer[yi][int(playerX/2)] & 12
			if singleByte != 0 {
				setCrashed()
			}
		}
	}
}

func checkCollisionBelow(oddEven float64, playerY float64, buffer [][]byte, playerX float64, onLaunchPad func(playerX float64, playerY float64) bool, setLanded func(), setCrashed func()) {
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
					setCrashed()
				}
			}
		} else {
			singleByte = buffer[yi][int(playerX/2)] & 3
			if singleByte != 0 {
				if singleByte == 3 && onLaunchPad(playerX, playerY) {
					setLanded()
				} else {
					setCrashed()
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
