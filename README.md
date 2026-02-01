# Golunar - Lunar Lander Game

A terminal-based Lunar Lander game written in Go using the tcell library. Land your spacecraft safely on the moon's surface while avoiding meteors and managing fuel efficiently!

Check your landing speed against actual speed.

## Features

- **Classic Lunar Lander Gameplay**: Navigate your lander to a safe landing while managing fuel consumption
- **Multiple Difficulty Levels**: Choose from Easy, Medium, and Hard modes
- **Meteor Avoidance**: Dodge incoming meteors to survive
- **Cross-Platform Support**: Runs on native platforms and in WebAssembly
- **Terminal UI**: Beautiful text-based graphics using tcell
- **Interactive Menu System**: Easy navigation between game modes and instructions

<a href="https://www.youtube.com/watch?feature=player_embedded&v=f8Lphefq4Hc" target="_blank">
 <img src="https://img.youtube.com/vi/f8Lphefq4Hc/0.jpg" alt="Watch the video" width="240" height="180" border="10" />
</a>

## Gameplay

### Controls

The controls are a compromise as you cannot detect key press and release events through the terminal. Hack is for thrust to be the up arrow key or shift and the left/right arrow key.

- **Left/Right Arrow Keys**: Rotate your lander
- **Up Arrow Key**: Fire main thruster
- **Shift + Left/Right Arrow**: Fire side thrusters
- **Enter/Escape**: Return to main menu

### Objective

- Land your spacecraft safely on the moon's surface
- Avoid collisions with meteors
- Conserve fuel to maximize your score
- Survive all difficulty levels

## Building

### Prerequisites

- Go 1.24.3 or higher
- Required dependencies: tcell/v3 and golang.org/x/image

### Native Build

```bash
go build -o golunar
./golunar
```

### WebAssembly Build

```bash
GOOS=js GOARCH=wasm go build -o wasm/main.wasm
cd wasm
python3 -m http.server
```
then visit http://127.0.0.1:8000

Then serve the `wasm/` directory using a web server to play in your browser.

## Project Structure

- `main.go` - Main game loop and initialization
- `landscape.go` - Landscape rendering and collision detection
- `meteor.go` - Meteor generation and movement logic
- `menu.go` - Interactive menu system
- `draw.go` - Drawing and rendering utilities
- `logo.go` - Game logo display
- `platform_native.go` - Native platform-specific code
- `platform_wasm.go` - WebAssembly platform-specific code
- `instructions.go` - In-game instructions display
- `wasm/` - WebAssembly web interface
  - `index.html` - Main HTML file
  - `wasm_exec.js` - JavaScript bridge for Go WASM
  - `tcell.js` - tcell terminal emulator for browser
  - `termstyle.css` - Terminal styling

## Dependencies

- [tcell/v3](https://github.com/gdamore/tcell) - Terminal/console toolkit
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) - Image manipulation

## Running the Game

After building, simply run the executable:

```bash
./golunar
```

The game will display a menu with the following options:
- Start Game Easy
- Start Game Medium
- Start Game Hard
- Instructions
- Exit

## License

See [LICENSE](LICENSE) file for details.

## Development

The project uses Go modules for dependency management. To install dependencies:

```bash
go mod download
```
