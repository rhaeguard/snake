package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
)

var bgColor = rl.NewColor(128, 160, 107, 255)
var snakeColor = rl.NewColor(0, 50, 44, 255)
var foodColor = rl.NewColor(0, 50, 44, 255)

const width = 1000
const height = 600
const step = 20

const (
	Up    int8 = -1
	Down       = 1
	Left       = -2
	Right      = 2
)

type Snake struct {
	pieces         [][]int32
	direction      int8
	lastUpdateTime float64
	stopped        bool
}

type Food struct {
	x, y           int32
	rotation       float32
	lastUpdateTime float64
}

var snake = Snake{
	pieces: [][]int32{
		{10, 6},
		{10, 7},
		{10, 8},
		{10, 9},
		{10, 10},
		{10, 11},
	},
	direction: Up,
	stopped:   true,
}

var food *Food = nil

var border = rl.NewRectangle(2*step, 2*step, width-4*step, height-6*step)
var borderThickness = float32(step) / 3

func randUInt32Between(min, max uint32) int32 {
	diff := max - min + 1
	u := rand.Uint32() % diff
	return int32(u + min)
}

func main() {
	rl.InitWindow(width, height, "retro snake")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	drawGrid := func() {
		//for x := 0; x <= width; x += step {
		//	rl.DrawLine(int32(x), 0, int32(x), height, rl.DarkGreen)
		//}
		//
		//for y := 0; y <= height; y += step {
		//	rl.DrawLine(0, int32(y), width, int32(y), rl.DarkGreen)
		//}

		// top
		rl.DrawRectangleV(rl.NewVector2(border.X, border.Y-borderThickness), rl.NewVector2(border.Width, borderThickness), rl.Black)
		// bottom
		rl.DrawRectangleV(rl.NewVector2(border.X, border.Y+border.Height), rl.NewVector2(border.Width, borderThickness), rl.Black)
		// left
		rl.DrawRectangleV(rl.NewVector2(border.X-borderThickness, border.Y-borderThickness), rl.NewVector2(borderThickness, border.Height+borderThickness*2), rl.Black)
		// right
		rl.DrawRectangleV(rl.NewVector2(border.X+border.Width, border.Y-borderThickness), rl.NewVector2(borderThickness, border.Height+borderThickness*2), rl.Black)
	}

	drawSnake := func() {
		for _, piece := range snake.pieces {
			x := piece[0]
			y := piece[1]

			r := rl.Rectangle{
				X:      float32(x * step),
				Y:      float32(y * step),
				Width:  step,
				Height: step,
			}
			rl.DrawRectangleRounded(r, 0.5, 100, snakeColor)
		}
	}

	nextHeadPosition := func(x, y int32) []int32 {
		switch snake.direction {
		case Left:
			x -= 1
		case Right:
			x += 1
		case Down:
			y += 1
		case Up:
			y -= 1
		}
		return []int32{x, y}
	}

	eatsItself := func(head []int32) bool {
		for _, piece := range snake.pieces {
			if piece[0] == head[0] && piece[1] == head[1] {
				return true
			}
		}

		return false
	}

	outOfBounds := func(sentinel []int32) bool {
		x := float32(sentinel[0] * step)
		y := float32(sentinel[1] * step)

		rHead := rl.NewRectangle(x, y, step, step)

		return !rl.CheckCollisionRecs(rHead, border)
	}

	updateSnake := func() {
		if rl.GetTime()-snake.lastUpdateTime < 0.125 {
			return
		}

		head := snake.pieces[0]
		x := head[0]
		y := head[1]

		newHeadPosition := nextHeadPosition(x, y)

		if outOfBounds(newHeadPosition) || eatsItself(newHeadPosition) {
			snake.stopped = true
			return
		}

		if food != nil {
			extendSnake := x == food.x && y == food.y
			if !extendSnake {
				snake.pieces = snake.pieces[:len(snake.pieces)-1]
			} else {
				food = nil
			}
		}

		snake.pieces = append([][]int32{newHeadPosition}, snake.pieces...)
		snake.lastUpdateTime = rl.GetTime()
	}

	grabKeyPresses := func() int8 {
		direction := snake.direction
		if rl.IsKeyPressed(rl.KeyLeft) {
			direction = Left
		}

		if rl.IsKeyPressed(rl.KeyRight) {
			direction = Right
		}

		if rl.IsKeyPressed(rl.KeyUp) {
			direction = Up
		}

		if rl.IsKeyPressed(rl.KeyDown) {
			direction = Down
		}

		if rl.IsKeyPressed(rl.KeySpace) {
			snake.stopped = !snake.stopped
		}

		if direction*-1 != snake.direction {
			snake.direction = direction
			return direction
		}

		return 0
	}

	addFood := func() {
		easeOut := func(t float64) float64 {
			return 1 - math.Pow(1-t, 3)
		}

		if food == nil {
			x := randUInt32Between(uint32(border.X/step), uint32((border.X+border.Width)/step)-1)
			y := randUInt32Between(uint32(border.Y/step), uint32((border.Y+border.Height)/step)-1)
			food = &Food{
				x:              x,
				y:              y,
				rotation:       720,
				lastUpdateTime: rl.GetTime(),
			}
		} else {
			t := rl.GetTime()
			progress := math.Min(1, (t-food.lastUpdateTime)/4)
			angle := 720 - 720*easeOut(progress)
			food.rotation = float32(angle)
		}
	}

	// basically rotates 4 points of the rectangle around origin
	// then draws two right-angle triangles
	drawRotatedRect := func(p rl.Rectangle, o rl.Vector2, color rl.Color) {
		rotatePtn := func(theta float64, p, o rl.Vector2) rl.Vector2 {
			px := math.Cos(theta)*float64(p.X-o.X) - math.Sin(theta)*float64(p.Y-o.Y) + float64(o.X)
			py := math.Sin(theta)*float64(p.X-o.X) + math.Cos(theta)*float64(p.Y-o.Y) + float64(o.Y)
			return rl.NewVector2(float32(px), float32(py))
		}

		theta := float64(food.rotation * (math.Pi / 180))

		ptl := rl.NewVector2(p.X, p.Y)
		ptr := rl.NewVector2(p.X+p.Width, p.Y)
		pbl := rl.NewVector2(p.X, p.Y+p.Height)
		pbr := rl.NewVector2(p.X+p.Width, p.Y+p.Height)

		ptl = rotatePtn(theta, ptl, o)
		ptr = rotatePtn(theta, ptr, o)
		pbl = rotatePtn(theta, pbl, o)
		pbr = rotatePtn(theta, pbr, o)

		rl.DrawTriangle(pbr, ptr, ptl, color)
		rl.DrawTriangle(pbr, ptl, pbl, color)
	}

	drawFood := func() {
		if food != nil {
			x := float32(food.x * step)
			y := float32(food.y * step)

			// drawing plus symbol
			w := float32(step) / 3
			h := float32(step)
			o := rl.NewVector2(x+w/2+w, y+h/2)

			vertical := rl.NewRectangle(x+w, y, w, h)
			horizontal := rl.NewRectangle(x, y+w, h, w)

			drawRotatedRect(vertical, o, foodColor)
			drawRotatedRect(horizontal, o, foodColor)
			rl.DrawCircleV(o, w/2, bgColor)
		}
	}

	for !rl.WindowShouldClose() {
		grabKeyPresses()
		addFood()
		if !snake.stopped {
			updateSnake()
		}

		rl.BeginDrawing()
		rl.ClearBackground(bgColor)
		// draw
		drawGrid()
		drawSnake()
		drawFood()
		rl.EndDrawing()
	}
}
