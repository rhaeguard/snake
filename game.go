package main

import (
	"fmt"
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var bgColor = rl.NewColor(128, 160, 107, 255)
var snakeColor = rl.NewColor(0, 50, 44, 255)
var foodColor = rl.NewColor(0, 50, 44, 255)

const width = 1280
const height = 720
const step = 20
const offsetX = 2
const offsetY = 2

// food rotation animation related
const rotationMax = 720
const totalRotateAnimationTime = 6

// game related
const maxPointsForFood = 20

// text related
const fontSize = 50
const textSpacing = 1.2

const (
	Up    int8 = -1
	Down  int8 = 1
	Left  int8 = -2
	Right int8 = 2
)

type Level = string

const (
	Level1 string = "SLUG"
	Level2 string = "WORM"
	Level3 string = "PYTHON"
)

var levels = []string{Level1, Level2, Level3}
var levelSpeed = map[string]float64{
	Level1: 0.125,
	Level2: 0.09,
	Level3: 0.0625,
}
var levelSkipScore = []struct {
	string
	uint32
}{
	{Level3, 100},
	{Level2, 50},
	{Level1, 0},
}

type Snake struct {
	pieces         [][]int32
	direction      int8
	lastUpdateTime float64
	score          uint32
	// game status
	started  bool
	paused   bool
	gameOver bool
	level    Level
}

type Food struct {
	x, y           int32
	rotation       float64
	lastUpdateTime float64
}

var food *Food = nil

var border = rl.NewRectangle(offsetX*step, offsetY*step, width-offsetX*2*step, height-offsetY*3*step)
var borderThickness = float32(step) / 3

type BorderDetails struct {
	top, bottom, left, right, horizontalThickness, verticalThickness rl.Vector2
}

var bd = BorderDetails{
	top:                 rl.NewVector2(border.X, border.Y-borderThickness),
	bottom:              rl.NewVector2(border.X, border.Y+border.Height),
	left:                rl.NewVector2(border.X-borderThickness, border.Y-borderThickness),
	right:               rl.NewVector2(border.X+border.Width, border.Y-borderThickness),
	horizontalThickness: rl.NewVector2(border.Width, borderThickness),
	verticalThickness:   rl.NewVector2(borderThickness, border.Height+borderThickness*2),
}

var foodRandXMin = uint32(border.X / step)
var foodRandXMax = uint32((border.X+border.Width)/step) - 1
var foodRandYMin = uint32(border.Y / step)
var foodRandYMax = uint32((border.Y+border.Height)/step) - 1

func randUInt32Between(min, max uint32) int32 {
	diff := max - min + 1
	u := rand.Uint32() % diff
	return int32(u + min)
}

var wfcPlane, startingPos = wfcInit(width/step-offsetX*2, height/step-offsetY*3)

var oceanAnimationLastUpdated = 0.0
var oceanAnimationFlip = false

var snake = Snake{
	pieces: [][]int32{
		{
			startingPos[1] + offsetX, // x pos
			startingPos[0] + offsetY, // y pos
		},
	},
	direction: Right,
	score:     0,
	paused:    true,
	started:   false,
	gameOver:  false,
	level:     Level1,
}

func main() {
	rl.InitWindow(width, height, "retro snake")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	font := rl.LoadFont("assets/Minecraft.ttf")
	defer rl.UnloadFont(font)

	var maxScore uint32 = 0

	drawGrid := func() {
		rl.DrawRectangleV(bd.top, bd.horizontalThickness, snakeColor)
		rl.DrawRectangleV(bd.bottom, bd.horizontalThickness, snakeColor)
		rl.DrawRectangleV(bd.left, bd.verticalThickness, snakeColor)
		rl.DrawRectangleV(bd.right, bd.verticalThickness, snakeColor)

		if snake.started {
			if rl.GetTime()-oceanAnimationLastUpdated > 0.6 {
				oceanAnimationFlip = !oceanAnimationFlip
				oceanAnimationLastUpdated = rl.GetTime()
			}

			for y, row := range wfcPlane {
				for x, tile := range row {
					if tile[0] == 'L' {
						// land
					} else if tile[0] == 'C' {
						// coast
						xp := float32((x + offsetX) * step)
						yp := float32((y + offsetY) * step)

						c := 4
						incr := float32(step) / float32(c)

						for ix := 1; ix < c; ix++ {
							xx := xp + float32(ix)*incr
							for iy := 1; iy < c; iy++ {
								yy := yp + float32(iy)*incr
								rl.DrawCircleV(rl.NewVector2(xx, yy), 1, rl.Black)
							}
						}

					} else {
						// sea
						xp := float32((x + offsetX) * step)
						yp := float32((y + offsetY) * step)

						if oceanAnimationFlip {
							xs := []float32{
								xp, xp + step/2, xp + step,
							}

							ys := []float32{
								yp + step/4, yp + (step - step/4), yp + step/4,
							}

							for i := 0; i < len(xs)-1; i++ {
								rl.DrawLineV(rl.NewVector2(xs[i], ys[i]), rl.NewVector2(xs[i+1], ys[i+1]), rl.Black)
							}
						} else {
							xs := []float32{
								xp, xp + (step / 4), xp + 3*(step/4), xp + step,
							}

							ys := []float32{
								yp + step/2, yp + step/4, yp + (step - step/4), yp + step/2,
							}

							for i := 0; i < len(xs)-1; i++ {
								rl.DrawLineV(rl.NewVector2(xs[i], ys[i]), rl.NewVector2(xs[i+1], ys[i+1]), rl.Black)
							}
						}
					}
				}
			}
		}
	}

	reflectAlongAxis := func(vertices []rl.Vector2, midpoint float32, xAxis bool) []rl.Vector2 {
		size := len(vertices)
		newVertices := make([]rl.Vector2, size)
		if xAxis {
			for ix, v := range vertices {
				diff := midpoint - v.X
				newVertices[size-ix-1] = rl.NewVector2(
					midpoint+diff,
					v.Y,
				)
			}
		} else {
			for ix, v := range vertices {
				diff := midpoint - v.Y
				newVertices[size-ix-1] = rl.NewVector2(
					v.X,
					midpoint+diff,
				)
			}
		}

		return newVertices
	}

	drawSnake := func() {
		for ix, piece := range snake.pieces {
			x := piece[0]
			y := piece[1]

			r := rl.Rectangle{
				X:      float32(x * step),
				Y:      float32(y * step),
				Width:  step,
				Height: step,
			}

			if !(ix > 0 && ix == len(snake.pieces)-1) {
				rl.DrawRectangleRounded(r, 0.5, 100, snakeColor)
			}

			if ix == 0 {
				/*
					to render the head, we basically render 2 right angled triangles with the background color.
					however, to figure out the alignment depending on the direction the snake is heading, we use reflection along an x or y coordinate.

					we need 2 triangles to represent the head. in our case these are t1 and t2.
					t1 represents a triangle that is top-left side of a left-moving snake's head.
							to produce the full head, we only need to flip t1 across y axis around the y midpoint of the block
					t2 represents a triangle that is top-left side of a up-moving snake's head.
							to produce the full head, we only need to flip t2 across x axis around the x midpoint of the block

					full formulas for producing 2 triangles:
						left moving:  [t1, Y(t1)]
						right moving: [X(t1), Y(X(t1))]
						up moving: 	  [t2, X(t2)]
						left moving:  [Y(t2), X(Y(t2))]
						:::: X(t) means to rotate around x midpoint; Y(t) means to rotate around y midpoint


					==============
					the same idea applies to eyes as well.
				*/

				// eyes
				eye1 := []rl.Vector2{
					rl.NewVector2(r.X+r.Width*0.25, r.Y+r.Height*0.25),
				}

				var eye2 []rl.Vector2

				// head
				t1 := []rl.Vector2{
					rl.NewVector2(r.X+r.Width*0.5, r.Y),
					rl.NewVector2(r.X, r.Y),
					rl.NewVector2(r.X, r.Y+r.Height*0.33),
				}

				t2 := []rl.Vector2{
					rl.NewVector2(r.X, r.Y),
					rl.NewVector2(r.X, r.Y+r.Height*0.5),
					rl.NewVector2(r.X+r.Width*0.33, r.Y),
				}

				switch snake.direction {
				case Down:
					eye2 = reflectAlongAxis(eye1, r.X+r.Width*0.5, true)

					t1 = reflectAlongAxis(t1, r.Y+r.Height*0.5, false)
					t2 = reflectAlongAxis(t1, r.X+r.Width*0.5, true)
				case Up:
					eye1 = reflectAlongAxis(eye1, r.Y+r.Height*0.5, false)
					eye2 = reflectAlongAxis(eye1, r.X+r.Width*0.5, true)

					t2 = reflectAlongAxis(t1, r.X+r.Width*0.5, true)
				case Right:
					eye2 = reflectAlongAxis(eye1, r.Y+r.Height*0.5, false)

					t1 = reflectAlongAxis(t1, r.X+r.Width*0.5, true)
					t2 = reflectAlongAxis(t1, r.Y+r.Height*0.5, false)
				case Left:
					eye1 = reflectAlongAxis(eye1, r.X+r.Width*0.5, true)
					eye2 = reflectAlongAxis(eye1, r.Y+r.Height*0.5, false)

					t2 = reflectAlongAxis(t1, r.Y+r.Height*0.5, false)
				}

				rl.DrawCircleV(eye1[0], 2.0, bgColor)
				rl.DrawCircleV(eye2[0], 2.0, bgColor)

				rl.DrawTriangle(t1[0], t1[1], t1[2], bgColor)
				rl.DrawTriangle(t2[0], t2[1], t2[2], bgColor)
			} else if ix > 0 && ix == len(snake.pieces)-1 {
				prev := snake.pieces[ix-1]
				px := prev[0]
				py := prev[1]
				var direction int8

				if px > x {
					direction = Right
				}
				if px < x {
					direction = Left
				}
				if py > y {
					direction = Down
				}
				if py < y {
					direction = Up
				}

				// tail
				var center rl.Vector2
				var startAngle float32
				switch direction {
				case Down, Right:
					center = rl.NewVector2(
						r.X+r.Width,
						r.Y+r.Height,
					)
					startAngle = 180.0
				case Up, Left:
					center = rl.NewVector2(
						r.X,
						r.Y,
					)
					startAngle = 0.0
				}

				rl.DrawCircleSector(
					center,
					r.Width,
					startAngle,
					startAngle+90.0,
					0,
					snakeColor,
				)
			}
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

	outOfBounds := func(head []int32) bool {
		x := float32(head[0] * step)
		y := float32(head[1] * step)

		rHead := rl.NewRectangle(x, y, step, step)

		return !rl.CheckCollisionRecs(rHead, border)
	}

	drowns := func(head []int32) bool {
		x := head[0]
		y := head[1]
		return wfcPlane[y-offsetY][x-offsetX][0] == 'S'
	}

	updateSnake := func() {
		if rl.GetTime()-snake.lastUpdateTime < levelSpeed[snake.level] {
			return
		}

		head := snake.pieces[0]
		x := head[0]
		y := head[1]

		newHeadPosition := nextHeadPosition(x, y)

		if outOfBounds(newHeadPosition) || eatsItself(newHeadPosition) || drowns(newHeadPosition) {
			snake.gameOver = true
			return
		}

		if food != nil {
			extendSnake := x == food.x && y == food.y
			if extendSnake {
				pct := food.rotation / rotationMax
				score := maxPointsForFood * pct
				snake.score += uint32(math.Max(1, score))
				if snake.score > maxScore {
					maxScore = snake.score
				}
				food = nil
			} else {
				snake.pieces = snake.pieces[:len(snake.pieces)-1]
			}
		}

		for _, v := range levelSkipScore {
			if maxScore > v.uint32 {
				if slices.Index(levels, snake.level) < slices.Index(levels, v.string) {
					snake.level = v.string
				}
				break
			}
		}

		snake.pieces = append([][]int32{newHeadPosition}, snake.pieces...)
		snake.lastUpdateTime = rl.GetTime()
	}

	grabKeyPresses := func() {
		if snake.started {
			direction := snake.direction
			if rl.IsKeyDown(rl.KeyLeft) {
				direction = Left
			}

			if rl.IsKeyDown(rl.KeyRight) {
				direction = Right
			}

			if rl.IsKeyDown(rl.KeyUp) {
				direction = Up
			}

			if rl.IsKeyDown(rl.KeyDown) {
				direction = Down
			}

			// don't allow moving in the opposite direction
			if direction*-1 != snake.direction {
				snake.direction = direction
			}
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			if snake.gameOver {
				wfcPlane, startingPos = wfcInit(width/step-4, height/step-6)
				snake = Snake{
					pieces: [][]int32{
						{
							startingPos[1] + offsetX, // x pos
							startingPos[0] + offsetY, // y pos
						},
					},
					direction: Right,
					score:     0,
					paused:    false,
					started:   true,
					gameOver:  false,
					level:     snake.level,
				}
				food = nil
			} else {
				snake.paused = !snake.paused
				snake.started = true
			}
		}

		if snake.gameOver && rl.IsKeyPressed(rl.KeySpace) {
			snake = Snake{
				pieces: [][]int32{
					{10, 6},
				},
				direction: Right,
				score:     0,
				paused:    true,
				started:   false,
				gameOver:  false,
				level:     snake.level,
			}
			food = nil
		}

		if !snake.started {
			var current int

			for i, l := range levels {
				if l == snake.level {
					current = i
					break
				}
			}

			if rl.IsKeyPressed(rl.KeyLeft) {
				current -= 1

				if current < 0 {
					current = len(levels) - 1
				}
			}

			if rl.IsKeyPressed(rl.KeyRight) {
				current += 1

				if current >= len(levels) {
					current = 0
				}
			}

			snake.level = levels[current]
		}

	}

	addFood := func() {
		generateNewFood := func() (int32, int32) {
		Selector:
			for {
				x := randUInt32Between(foodRandXMin, foodRandXMax)
				y := randUInt32Between(foodRandYMin, foodRandYMax)

				if wfcPlane[y-offsetY][x-offsetX][0] == 'S' {
					continue Selector
				}

				for _, piece := range snake.pieces {
					if piece[0] == x && piece[1] == y {
						continue Selector
					}
				}

				return x, y
			}
		}

		easeOut := func(t float64) float64 {
			return 1 - math.Pow(1-t, 3)
		}

		if food != nil && rl.GetTime()-food.lastUpdateTime >= 10 {
			food = nil
		}

		if food == nil {
			x, y := generateNewFood()
			food = &Food{
				x:              x,
				y:              y,
				rotation:       rotationMax,
				lastUpdateTime: rl.GetTime(),
			}
		} else {
			progress := math.Min(1, (rl.GetTime()-food.lastUpdateTime)/totalRotateAnimationTime)
			food.rotation = rotationMax * (1 - easeOut(progress))
		}
	}

	// rotates the point p around point o by theta radians
	rotatePtn := func(theta float64, p, o rl.Vector2) rl.Vector2 {
		cos := float32(math.Cos(theta))
		sin := float32(math.Sin(theta))
		dx := p.X - o.X
		dy := p.Y - o.Y
		px := cos*dx - sin*dy + o.X
		py := sin*dx + cos*dy + o.Y
		return rl.NewVector2(px, py)
	}

	// basically rotates 4 points of the rectangle around origin
	// then draws two right-angle triangles
	drawRotatedRect := func(p rl.Rectangle, o rl.Vector2, color rl.Color) {
		theta := food.rotation * rl.Deg2rad

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
			// location of the food cell
			x := float32(food.x * step)
			y := float32(food.y * step)

			// drawing plus symbol
			// width
			w := float32(step) / 3
			// height
			h := float32(step)
			// center
			o := rl.NewVector2(x+w/2+w, y+h/2)

			// vertical and horizontal rectangles that make up the + symbol
			vertical := rl.NewRectangle(x+w, y, w, h)
			horizontal := rl.NewRectangle(x, y+w, h, w)

			drawRotatedRect(vertical, o, foodColor)
			drawRotatedRect(horizontal, o, foodColor)
			rl.DrawCircleV(o, w/2, bgColor) // it's in bgColor to imitate hollowness
		}
	}

	drawHud := func() {
		text := fmt.Sprintf("SCORE : %d/%d", snake.score, maxScore)
		position := rl.NewVector2(border.X, border.Y+border.Height+20)
		rl.DrawTextEx(font, text, position, fontSize, textSpacing, snakeColor)

		text = fmt.Sprintf("LVL : %s", snake.level)
		size := rl.MeasureTextEx(font, text, fontSize, textSpacing)
		position = rl.NewVector2(border.X+border.Width-size.X, border.Y+border.Height+20)
		rl.DrawTextEx(font, text, position, fontSize, textSpacing, snakeColor)
	}

	drawCenteredText := func(text ...string) float32 {
		fullTextHeight := float32(len(text) * fontSize)
		for i, t := range text {
			size := rl.MeasureTextEx(font, t, fontSize, textSpacing)
			xx := (width - size.X) / 2
			yy := (height-fullTextHeight)/2 + float32(i*fontSize)

			position := rl.NewVector2(xx, yy)
			rl.DrawTextEx(font, t, position, fontSize, textSpacing, snakeColor)
		}

		return (height-fullTextHeight)/2 + float32((len(text)+1)*fontSize)
	}

	drawCenteredTextFromPosition := func(posY float32, options ...string) {
		const y = width * 0.05

		var K float32

		for _, o := range options {
			K += rl.MeasureTextEx(font, o, fontSize, textSpacing).X
		}

		var X float32
		X = 0.5 * (width - 2*y - K)

		var prefix = X

		padding := func(position, size rl.Vector2) {
			newPosition := rl.NewVector2(position.X-10, position.Y-10)
			newSize := rl.NewVector2(size.X+10, size.Y+10)
			rl.DrawRectangleV(newPosition, newSize, snakeColor)
		}

		for i, o := range options {
			size := rl.MeasureTextEx(font, o, fontSize, textSpacing)
			position := rl.NewVector2(prefix, posY)
			if o == snake.level {
				padding(position, size)
				rl.DrawTextEx(font, o, position, fontSize, textSpacing, bgColor)
			} else {
				rl.DrawTextEx(font, o, position, fontSize, textSpacing, snakeColor)
			}
			prefix = prefix + size.X
			if i != len(options)-1 {
				prefix += y
			}
		}
	}

	drawGameTitle := func(t string) {
		size := rl.MeasureTextEx(font, t, 100, textSpacing)
		xx := (width - size.X) / 2
		yy := height * 0.2

		position := rl.NewVector2(xx, float32(yy))
		rl.DrawTextEx(font, t, position, 100, textSpacing, snakeColor)
	}

	for !rl.WindowShouldClose() {
		grabKeyPresses()
		addFood()
		if !snake.paused {
			updateSnake()
		}

		rl.BeginDrawing()
		rl.ClearBackground(bgColor)
		drawGrid()
		// draw
		if snake.started && !snake.gameOver {
			drawSnake()
			drawFood()
			drawHud()
		} else if snake.gameOver {
			drawCenteredText("GAME OVER", "ENTER TO RESTART", "SPACE TO MENU")
		} else {
			drawGameTitle(".....SNAKE.....")
			py := drawCenteredText("PRESS ENTER TO START")
			drawCenteredTextFromPosition(py, levels...)
		}

		rl.EndDrawing()
	}

}
