package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
)

//go:embed main.go
var source string

var (
	font         rl.Font
	fontSize     = float32(20)
	screenWidth  = float32(600)
	screenHeight = float32(400)
)

func main() {
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "raylib [core] example - basic window")
	defer rl.CloseWindow()
	rl.SetWindowPosition(800, 350)
	rl.SetWindowState(rl.FlagWindowResizable)

	font = rl.LoadFontEx("/Users/philipp/Library/Fonts/JetBrainsMono-Medium.ttf", 40, nil, 0)

	source := strings.TrimSpace(strings.ReplaceAll(source, "\t", "    "))
	sourceLines := strings.Split(source, "\n")
	sourceOffset := float32(0)
	sourceHeight := float32(len(sourceLines)) * fontSize

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			screenWidth = float32(rl.GetScreenWidth())
			screenHeight = float32(rl.GetScreenHeight())
			fmt.Println("resized")
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		DrawLines(sourceLines, 0, sourceOffset, 20, rl.White)
		DrawMouseLine(sourceOffset)
		DrawTable(screenWidth-100, 0, 100, screenHeight)

		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			log.Println("pressed", GetMouseLinePosition(sourceOffset))
		}

		if move := rl.GetMouseWheelMove(); move != 0 {
			sourceOffset += move * 10
			sourceOffset = clip(sourceOffset, -sourceHeight+screenHeight, 0)

			// log.Println("wheel", rl.GetMouseWheelMove())
		}

		rl.DrawFPS(int32(screenWidth)-100, 0)
		rl.EndDrawing()
	}
}

func DrawTable(x, y, w, h float32) {
	rl.DrawRectangleV(rl.Vector2{X: x, Y: y}, rl.Vector2{X: w, Y: h}, rl.Black)

	rec := rl.Rectangle{X: x, Y: y, Width: w, Height: h}
	rl.DrawRectangleLinesEx(rec, 1, rl.DarkGreen)

	numRows := int(h / fontSize)
	for i := 1; i < numRows; i++ {
		start := rl.Vector2{X: x, Y: y + (float32(i) * fontSize)}
		end := rl.Vector2{X: x + w, Y: y + (float32(i) * fontSize)}
		rl.DrawLineEx(start, end, 1, rl.DarkGreen)
	}
}

func GetMouseLinePosition(offset float32) int32 {
	pos := rl.GetMousePosition()
	return int32((pos.Y - offset) / fontSize)
}

func DrawMouseLine(offset float32) {
	lineIndex := GetMouseLinePosition(offset)
	recY := float32(lineIndex)*fontSize + offset
	rec := rl.Rectangle{X: 0, Y: recY, Width: float32(screenWidth), Height: 20}
	rl.DrawRectangleLinesEx(rec, 1, rl.DarkGreen)
}

func DrawLines(lines []string, x, y, fontSize float32, color color.RGBA) {
	var offset float32
	for _, line := range lines {
		DrawLine(line, x, y+offset, fontSize, color)
		offset += fontSize
	}
}

func DrawLine(text string, x, y, fontSize float32, color color.RGBA) {
	rl.DrawTextEx(font, text, rl.Vector2{X: x, Y: y}, fontSize, 0, color)
}

func clip[T constraints.Ordered](v, min, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
