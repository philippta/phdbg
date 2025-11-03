package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed main.go
var source string

var tableData = [][]string{
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
	{"foo", "100", "int32"},
}

var (
	DarkerBlue  = rl.NewColor(0, 65, 137, 255)
	AlmostBlack = rl.NewColor(13, 20, 35, 255)

	font            rl.Font
	fontSize        = float32(20)
	screenWidth     = float32(756)
	screenHeight    = float32(916)
	screenPositionX = 756
	screenPositionY = 66
	// screenWidth     = float32(1280)
	// screenHeight    = float32(1387)
	// screenPositionX = 1280
	// screenPositionY = 53
	charWidth = float32(0)
)

func main() {
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "raylib [core] example - basic window")
	defer rl.CloseWindow()

	rl.SetWindowPosition(screenPositionX, screenPositionY)
	rl.SetWindowState(rl.FlagWindowResizable)
	// rl.EnableEventWaiting()

	font = rl.LoadFontEx("/Users/philipp/Library/Fonts/JetBrainsMono-Medium.ttf", int32(fontSize*2), nil, 0)

	source := strings.TrimSpace(strings.ReplaceAll(source, "\t", "    "))
	sourceScroll := rl.Vector2{}
	charWidth = rl.MeasureTextEx(font, " ", fontSize, 0).X

	rt := DrawTextToTexture(source)
	defer rl.UnloadRenderTexture(rt)

	for !rl.WindowShouldClose() {
		mouse := rl.GetMousePosition()
		if rl.IsWindowResized() {
			sourceScroll = rl.Vector2{}
			DebugWindowSize()
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		var x, y, w, h float32

		{
			// Source
			x, y, w, h = 0, 0, screenWidth/2, screenHeight*3/4

			if InRegion(x, y, w, h, mouse) {
				mov := rl.GetMouseWheelMoveV()
				sourceScroll.Y -= mov.Y * 10
				sourceScroll.X -= mov.X * 10

				sourceScroll.Y = min(max(0, sourceScroll.Y), float32(rt.Texture.Height)-h)
				sourceScroll.X = min(max(0, sourceScroll.X), float32(rt.Texture.Width)-w)
			}
			DrawTextureClipped(x, y+fontSize, w, h, sourceScroll.X, sourceScroll.Y, rt)

			// Scroll Indicator
			scrollerHeight := float32(24)
			scrollerWidth := float32(4)
			scrollerPos := ((y + sourceScroll.Y) / (float32(rt.Texture.Height) - h)) * (h - scrollerHeight - fontSize)
			DrawBackground(w-scrollerWidth, scrollerPos+fontSize, scrollerWidth, scrollerHeight, rl.DarkBlue)
			DrawOutline(x, y, w, h, rl.DarkBlue)

			DrawBackground(x, y, w, fontSize, DarkerBlue)
			DrawText(x+4, y, w, "main.go", rl.White)
			DrawHLine(x, y+fontSize, w, rl.DarkBlue)
		}
		{
			// Table
			x += w - 1
			w = screenWidth - x
			DrawTable(x, y, w, h, tableData, rl.White)
			DrawOutline(x, y, w, h, rl.DarkBlue)
		}
		{
			// Bottom Box
			x = 0
			y = h
			w = screenWidth
			h = screenHeight - h
			DrawBackground(x, y, w, h, rl.Black)
			DrawOutline(x, y, w, h, rl.DarkBlue)
		}

		rl.DrawFPS(int32(screenWidth)-100, int32(screenHeight)-20)
		rl.EndDrawing()
	}
}

func DrawTextureClipped(x, y, w, h, offsetX, offsetY float32, rt rl.RenderTexture2D) {
	tw, th := float32(rt.Texture.Width), float32(rt.Texture.Height)
	src := rl.Rectangle{X: offsetX, Y: th + offsetY, Width: min(tw-offsetX, w), Height: min(th-offsetY, h)}
	dst := rl.Rectangle{X: x, Y: y, Width: min(tw-offsetX, w), Height: min(th-offsetY, h)}
	origin := rl.Vector2{X: 0, Y: 0}
	rl.DrawTexturePro(rt.Texture, src, dst, origin, 0, rl.White)
}

func DrawTextToTexture(text string) rl.RenderTexture2D {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\t", "    "))
	lines := strings.Split(text, "\n")
	charsize := rl.MeasureTextEx(font, " ", fontSize, 0)

	longestLine := 0
	for _, line := range lines {
		longestLine = max(longestLine, len(line))
	}

	height := charsize.Y * float32(len(lines))
	width := charsize.X*float32(longestLine) + 8

	rt := rl.LoadRenderTexture(int32(width*2), int32(height*2))
	rl.SetTextureWrap(rt.Texture, rl.WrapMirrorRepeat)

	rl.BeginTextureMode(rt)
	DrawBackground(0, 0, float32(rt.Texture.Width), float32(rt.Texture.Height), rl.Black)
	rowY := float32(0)
	for _, line := range lines {
		rl.DrawTextEx(font, line, rl.Vector2{X: 8, Y: rowY}, fontSize*2, 0, rl.White)
		rowY += fontSize * 2
	}

	rl.EndTextureMode()

	rt.Texture.Width /= 2
	rt.Texture.Height /= 2

	return rt
}

func DrawTable(x, y, w, h float32, rows [][]string, color color.RGBA) {
	if len(rows) == 0 {
		return
	}

	numCols := len(rows[0])
	colWidth := w / float32(numCols)
	numRows := int(h / fontSize)

	rowY := y
	rowX := x + 4
	DrawBackground(x, rowY, w, fontSize, DarkerBlue)
	DrawText(rowX, y, w, "Name", rl.White)
	rowX += colWidth
	DrawText(rowX, y, w, "Value", rl.White)
	rowX += colWidth
	DrawText(rowX, y, w, "Type", rl.White)
	rowY += fontSize
	DrawHLine(x, rowY, w, rl.DarkBlue)

	for i := range numRows {
		rowX := x + 4
		if i%2 == 1 {
			DrawBackground(x, rowY, w, fontSize, AlmostBlack)
		}

		if i < len(rows) {
			row := rows[i]
			for _, text := range row {
				text = clips(text, colWidth)
				DrawText(rowX, rowY, w, text, rl.White)
				rowX += colWidth
			}
		}
		rowY += fontSize
	}
}

func DrawMultilineText(x, y, w, h float32, lines []string, color color.RGBA) {
	rowY := y
	for _, line := range lines {
		DrawText(x, rowY, w, line, rl.White)
		rowY += fontSize
		if rowY >= h {
			break
		}
	}
}

func DrawBackground(x, y, w, h float32, color color.RGBA) {
	rl.DrawRectangleV(rl.Vector2{X: x, Y: y}, rl.Vector2{X: w, Y: h}, color)
}

func DrawOutline(x, y, w, h float32, color color.RGBA) {
	rec := rl.Rectangle{X: x, Y: y, Width: w, Height: h}
	rl.DrawRectangleLinesEx(rec, 1, color)
}

func DrawText(x, y, w float32, text string, color color.RGBA) {
	rl.DrawTextEx(font, clips(text, w), rl.Vector2{X: x, Y: y}, fontSize, 0, color)
}

func DrawHLine(x, y, w float32, color color.RGBA) {
	start := rl.Vector2{X: x, Y: y}
	end := rl.Vector2{X: x + w, Y: y}
	rl.DrawLineEx(start, end, 1, color)
}

func DrawVLine(x, y, h float32, color color.RGBA) {
	start := rl.Vector2{X: x, Y: y}
	end := rl.Vector2{X: x, Y: y + h}
	rl.DrawLineEx(start, end, 1, color)
}

func InRegion(x, y, w, h float32, pos rl.Vector2) bool {
	if pos.X >= x && pos.X < x+w {
		if pos.Y >= y && pos.Y < y+h {
			return true
		}
	}
	return false
}

func DebugWindowSize() {
	screenWidth = float32(rl.GetScreenWidth())
	screenHeight = float32(rl.GetScreenHeight())
	pos := rl.GetWindowPosition()

	fmt.Printf(`
	screenWidth     = float32(%d)
	screenHeight    = float32(%d)
	screenPositionX = %d
	screenPositionY = %d
	`, int(screenWidth), int(screenHeight), int(pos.X), int(pos.Y))
}

func clips(s string, w float32) string {
	maxc := int(w / charWidth)
	if len(s) > maxc {
		return s[:maxc]
	}
	return s
}

func abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}
