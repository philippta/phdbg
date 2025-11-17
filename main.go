package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"
	"sync/atomic"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/rpc2"
)

var tableData = [][3]string{
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

	font      rl.Font
	fontSize  = float32(20)
	charWidth = float32(0)

	screenWidth     = float32(756)
	screenHeight    = float32(916)
	screenPositionX = 756
	screenPositionY = 66

	// screenWidth     = float32(1280)
	// screenHeight    = float32(1387)
	// screenPositionX = 1280
	// screenPositionY = 53
)

// bounds
var (
	sourceBounds = rl.Rectangle{X: 0, Y: 0, Width: screenWidth * 0.5, Height: screenHeight}
	watchBounds  = rl.Rectangle{X: sourceBounds.Width - 1, Y: 0, Width: screenWidth - sourceBounds.Width + 1, Height: screenHeight}
	boxBounds    = rl.Rectangle{X: 0, Y: sourceBounds.Height - 1, Width: screenWidth, Height: screenHeight - sourceBounds.Height + 1}
)

func updateBounds() {
	sourceBounds = rl.Rectangle{X: 0, Y: 0, Width: screenWidth * 0.5, Height: screenHeight}
	watchBounds = rl.Rectangle{X: sourceBounds.Width - 1, Y: 0, Width: screenWidth - sourceBounds.Width + 1, Height: screenHeight}
	// boxBounds = rl.Rectangle{X: 0, Y: sourceBounds.Height - 1, Width: screenWidth, Height: screenHeight - sourceBounds.Height + 1}
}

var (
	debugger      *rpc2.RPCClient
	debuggerState atomic.Pointer[api.DebuggerState]
)

func main() {
	// cmd := exec.Command("dlv", "--headless", "--listen", "127.0.0.1:6060", "exec", "./foo/foo")
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	// cmd.Start()

	debugger = rpc2.NewClient("127.0.0.1:6060")
	LoadDebugState()
	debugger.CreateBreakpoint(&api.Breakpoint{FunctionName: "main.main"})

	<-debugger.Continue()
	LoadDebugState()
	LoadVariables()

	rl.SetConfigFlags(rl.FlagWindowUnfocused)
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "phdbg")
	defer rl.CloseWindow()

	rl.SetWindowPosition(screenPositionX, screenPositionY)
	rl.SetWindowState(rl.FlagWindowResizable)
	// rl.SetTargetFPS(60)
	// rl.EnableEventWaiting()

	font = rl.LoadFontEx("/Users/philipp/Library/Fonts/JetBrainsMono-Medium.ttf", int32(fontSize*2), nil, 0)
	charWidth = rl.MeasureTextEx(font, " ", fontSize, 0).X

	_ = sourceScroll

	for !rl.WindowShouldClose() {
		mouse := rl.GetMousePosition()
		_ = mouse
		if rl.IsWindowResized() {
			sourceScroll = rl.Vector2{}
			DebugWindowSize()
		}

		if rl.IsKeyPressed(rl.KeyN) {
			UpdateDebugState(debugger.Next())
			LoadVariables()
		}
		if rl.IsKeyPressed(rl.KeyS) {
			UpdateDebugState(debugger.Step())
			LoadVariables()
		}

		debugState := GetDebugState()
		sourceScroll.Y = -sourceBounds.Height/2 + float32(debugState.CurrentThread.Line-1)*fontSize

		tableData = [][3]string{}
		for _, vv := range localVariables {
			for _, v := range vv {
				tableData = append(tableData, [3]string{
					v.Name,
					v.Value,
					v.Type,
				})
			}
			tableData = append(tableData, [3]string{})
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		{
			// Source
			x, y, w, h := sourceBounds.X, sourceBounds.Y, sourceBounds.Width, sourceBounds.Height
			titleHeight := fontSize

			// if InRegion(x, y, w, h, mouse) {
			// 	mov := rl.GetMouseWheelMoveV()
			// 	sourceScroll.Y -= mov.Y * 10
			// 	sourceScroll.Y = min(max(0, sourceScroll.Y), sourceHeight-(h-titleHeight))
			// }

			// Text
			DrawMultilineText(x, y+titleHeight, w, h-titleHeight, -sourceScroll.Y, sourceLines, rl.White)
			DrawOutline(x, y+titleHeight, w, h-titleHeight, rl.DarkBlue)

			DrawOutline(x, sourceBounds.Height/2+titleHeight, w, fontSize, rl.DarkBlue)

			// Title
			DrawBackground(x, y, w, titleHeight, rl.DarkBlue)
			DrawText(x+4, y, w, "main.go", rl.White)

			// Scroll
			scrollPos := sourceScroll.Y / (sourceHeight - h + titleHeight)
			DrawScroller(w, y+titleHeight, h-titleHeight, scrollPos)
		}
		{

			// Table
			x, y, w, h := watchBounds.X, watchBounds.Y, watchBounds.Width, watchBounds.Height
			DrawTable(x, y, w, h, tableData, rl.White)
			DrawOutline(x, y, w, h, rl.DarkBlue)
		}
		{
			// Bottom Box
			// x, y, w, h := boxBounds.X, boxBounds.Y, boxBounds.Width, boxBounds.Height
			// DrawBackground(x, y, w, h, rl.Black)
			// DrawOutline(x, y, w, h, rl.DarkBlue)
		}

		rl.DrawFPS(int32(screenWidth)-100, int32(screenHeight)-20)
		// PrintDebugState(0, 0, screenWidth, screenHeight)
		rl.EndDrawing()
	}
}

func DrawScroller(x, y, h, pos float32) {
	scrollerHeight := float32(24)
	scrollerWidth := float32(4)
	rng := h - scrollerHeight
	DrawBackground(x-scrollerWidth, y+rng*pos, scrollerWidth, scrollerHeight, rl.DarkBlue)
}

func DrawTable(x, y, w, h float32, rows [][3]string, color color.RGBA) {
	numCols := 3
	colWidth := w / float32(numCols)

	rowY := y
	rowX := x + 4
	DrawBackground(x, rowY, w, fontSize, rl.DarkBlue)
	DrawText(rowX, y, w, "Name", rl.White)
	rowX += colWidth
	DrawText(rowX, y, w, "Value", rl.White)
	rowX += colWidth
	DrawText(rowX, y, w, "Type", rl.White)
	rowY += fontSize
	DrawHLine(x, rowY, w, rl.DarkBlue)

	if len(rows) == 0 {
		return
	}

	for _, row := range rows {
		rowX := x + 4
		if row[0] == "" && row[1] == "" && row[2] == "" {
			DrawHLine(rowX, rowY, w, rl.DarkBlue)
			continue
		}

		for _, text := range row {
			text = clips(text, colWidth)
			DrawText(rowX, rowY, w, text, rl.White)
			rowX += colWidth
		}
		rowY += fontSize
	}
}

func DrawMultilineText(x, y, w, h, scrollY float32, lines []string, color color.RGBA) {
	rowY := y
	for _, line := range lines {
		if rowY+scrollY+fontSize >= y {
			DrawText(x, rowY+scrollY, w, line, color)
		}
		rowY += fontSize
		if rowY+scrollY-fontSize >= h {
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
	updateBounds()

	fmt.Printf(`
	screenWidth     = float32(%d)
	screenHeight    = float32(%d)
	screenPositionX = %d
	screenPositionY = %d
	`, int(screenWidth), int(screenHeight), int(pos.X), int(pos.Y))
}

func xywh(x, y, w, h float32) (int32, int32, int32, int32) {
	return int32(x), int32(y), int32(h), int32(w)
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

var (
	perfLabel string
	perfStart time.Time
)

func BeginPerformanceMeasure(label string) {
	perfLabel = label
	perfStart = time.Now()
}

func EndPerformanceMeasure() {
	fmt.Printf(perfLabel+": %dns\n", time.Since(perfStart).Nanoseconds())
}

func BeginScissorMode(x, y, w, h float32) {
	rl.BeginScissorMode(int32(x), int32(y), int32(w), int32(h))
}

func EndScissorMode() {
	rl.EndScissorMode()
}

var (
	lastLoadedFile string
	sourceLines    []string
	sourceHeight   float32
	sourceScroll   rl.Vector2
	sourceWidth    float32
)

func LoadFile(name string) {
	if name == lastLoadedFile {
		return
	}

	data, err := os.ReadFile(name)
	if err != nil {
		return
	}
	source := string(data)
	source = strings.TrimSpace(strings.ReplaceAll(source, "\t", "    "))
	sourceLines = strings.Split(source, "\n")
	sourceHeight = float32(len(sourceLines)) * fontSize
	sourceScroll = rl.Vector2{}
	sourceWidth = float32(0)
	for _, l := range sourceLines {
		sourceWidth = max(sourceWidth, float32(len(l))*fontSize)
	}
}

var emptyDebuggerState = &api.DebuggerState{
	CurrentThread:     &api.Thread{Function: &api.Function{}},
	SelectedGoroutine: &api.Goroutine{},
}

func LoadDebugState() {
	UpdateDebugState(debugger.GetState())
}

func UpdateDebugState(s *api.DebuggerState, err error) {
	if err != nil {
		return
	}
	debuggerState.Store(s)
	if s.CurrentThread != nil {
		LoadFile(s.CurrentThread.File)
		fmt.Println(s.CurrentThread.File, s.CurrentThread.Line)
	}
}

func GetDebugState() *api.DebuggerState {
	s := debuggerState.Load()
	if s == nil {
		return emptyDebuggerState
	}
	return s
}

var normalLoadConfig = api.LoadConfig{
	FollowPointers:     true,
	MaxVariableRecurse: 1,
	MaxStringLen:       64,
	MaxArrayValues:     64,
	MaxStructFields:    -1,
}

var localVariables [][]api.Variable

func LoadVariables() {
	s := GetDebugState()
	if s.SelectedGoroutine.ID == 0 {
		return
	}
	var newVars [][]api.Variable

	for i := 10; i >= 0; i-- {
		var newVarsInner []api.Variable
		evalScope := api.EvalScope{Frame: i, GoroutineID: s.SelectedGoroutine.ID}

		if vars, err := debugger.ListFunctionArgs(evalScope, normalLoadConfig); err == nil {
			newVarsInner = append(newVarsInner, vars...)
		}
		if vars, err := debugger.ListLocalVariables(evalScope, normalLoadConfig); err == nil {
			newVarsInner = append(newVarsInner, vars...)
		}

		if len(newVarsInner) > 0 {
			newVars = append(newVars, newVarsInner)
		}
	}

	localVariables = newVars
}

func PrintDebugState(x, y, w, h float32) {
	state := debuggerState.Load()
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	DrawMultilineText(x, y, w, h, 0, lines, rl.Red)
}
