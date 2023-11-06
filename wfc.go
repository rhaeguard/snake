package main

import (
	"fmt"
	"math"
	"math/rand"
)

var inputMatrix = [][]uint8{
	{'L', 'L', 'L', 'L', 'L'},
	{'L', 'L', 'L', 'L', 'L'},
	{'L', 'C', 'C', 'C', 'L'},
	{'C', 'S', 'S', 'S', 'C'},
	{'S', 'S', 'S', 'S', 'S'},
	{'S', 'S', 'S', 'S', 'S'},
	{'S', 'S', 'S', 'S', 'S'},
}

type v2 struct {
	x, y int
}

func validDirections(mh, mw, x, y int) []v2 {
	var ds []v2

	// up
	if y-1 >= 0 {
		ds = append(ds, v2{0, -1})
	}
	// down
	if y+1 < mh {
		ds = append(ds, v2{0, 1})
	}
	// left
	if x-1 >= 0 {
		ds = append(ds, v2{-1, 0})
	}
	// right
	if x+1 < mw {
		ds = append(ds, v2{1, 0})
	}

	return ds
}

func generateRules(matrix [][]uint8) (map[string]bool, map[uint8]uint) {
	rules := make(map[string]bool)
	weights := make(map[uint8]uint)
	for y, row := range matrix {
		for x, tile := range row {
			if _, ok := weights[tile]; !ok {
				weights[tile] = 0
			}
			weights[tile] += 1
			directions := validDirections(len(matrix), len(matrix[0]), x, y)
			for _, d := range directions {
				a := matrix[y+d.y][x+d.x]
				k := fmt.Sprintf("%c%c%d%d", tile, a, d.x, d.y)
				rules[k] = true
			}
		}
	}

	return rules, weights
}

func getLowestEntropyCoords(weights map[uint8]uint, plane [][][]uint8) v2 {
	shannonEntropy := func(options []uint8) float64 {
		sm := 0.0
		smLog := 0.0
		for _, o := range options {
			ww := float64(weights[o])
			sm += ww
			smLog += ww * math.Log(ww)
		}
		return math.Log(sm) - (smLog / sm)
	}
	min := math.Inf(1)
	var coords = v2{}
	for y, row := range plane {
		for x, options := range row {
			if len(options) == 1 {
				continue
			}

			e := shannonEntropy(options)
			e = e - (rand.Float64() / 1000)
			if e < min {
				min = e
				coords = v2{x, y}
			}
		}
	}

	return coords
}

func collapse(coords v2, weights map[uint8]uint, plane [][][]uint8) {
	opts := plane[coords.y][coords.x]

	totalWeight := 0.0
	for _, o := range opts {
		totalWeight += float64(weights[o])
	}

	totalWeight = totalWeight * rand.Float64()

	pick := opts[0]

	for _, o := range opts {
		totalWeight -= float64(weights[o])
		if totalWeight < 0 {
			pick = o
			break
		}
	}
	//pick := opts[rand.Intn(len(opts))]
	plane[coords.y][coords.x] = []uint8{pick}
}

func propagate(coords v2, rules map[string]bool, plane [][][]uint8) {
	stack := []v2{coords}

	for len(stack) != 0 {
		curCoords := stack[len(stack)-1]
		stack = stack[0 : len(stack)-1]

		tiles := plane[curCoords.y][curCoords.x]
		ds := validDirections(len(plane), len(plane[0]), curCoords.x, curCoords.y)

		for _, d := range ds {
			options := plane[curCoords.y+d.y][curCoords.x+d.x]
			var keep []uint8
			for _, otherTile := range options {
				var ok bool
				for _, tile := range tiles {
					k := fmt.Sprintf("%c%c%d%d", tile, otherTile, d.x, d.y)
					if _, ok = rules[k]; ok {
						break
					}
				}

				if ok {
					keep = append(keep, otherTile)
				} else {
					stack = append(stack, v2{curCoords.x + d.x, curCoords.y + d.y})
				}
			}
			if keep != nil {
				plane[curCoords.y+d.y][curCoords.x+d.x] = keep
			}
		}
	}
}

func fullyCollapsed(plane [][][]uint8) bool {
	for _, row := range plane {
		for _, opts := range row {
			if len(opts) != 1 {
				return false
			}
		}
	}

	return true
}

func wfc(weights map[uint8]uint, rules map[string]bool, w, h int) [][][]uint8 {
	// init plane - start
	var plane [][][]uint8 = make([][][]uint8, h)

	for yy := 0; yy < h; yy++ {
		plane[yy] = make([][]uint8, w)
		for xx := 0; xx < w; xx++ {
			plane[yy][xx] = []uint8{'L', 'C', 'S'}
		}
	}
	// init plane - end

	for !fullyCollapsed(plane) {
		c := getLowestEntropyCoords(weights, plane)
		collapse(c, weights, plane)
		propagate(c, rules, plane)
	}

	return plane
}

func wfcInit(w, h int) [][][]uint8 {
	rules, weights := generateRules(inputMatrix)
	return wfc(weights, rules, w, h)
}
