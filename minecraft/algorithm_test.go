package minecraft_test

import (
	"fmt"
	"phoenix/minecraft/function"
	"phoenix/minecraft/function/fetcher"
	"testing"
)

func TestCircle(t *testing.T){
	height := 5.0
	radius := 5.0
	var vec []function.Vector
	for h := 0.0; h <= height ; h += 1.0 {
		for x := -radius ; x <= radius ; x += 1.0 {
			for y := -radius ; y < radius ; y += 1.0 {
				if radius * radius > x * x + y * y && x * x + y * y >= (radius - 1) * (radius - 1) {
					vec = append(vec, []float64{h, x, y})
				}
			}
		}
	}

	Radius := 5
	Height := 5
	var v2 [][]int
	for h1 := 0; h1 <= Height; h1 += 1 {
		for i := -Radius; i <= Radius; i++ {
			for j := -Radius; j <= Radius; j++ {
				if i*i+j*j < Radius*Radius && i*i+j*j >= (Radius-1)*(Radius-1) {
					v2 = append(v2, []int{h1, i, j})
				}
			}
		}
	}

	fmt.Println(vec)
	fmt.Println(v2)
	fmt.Printf("%d %d", len(vec), len(v2))
}

const Unit = 100
func SplitArea(length float64) float64 {
	curr := 1.0
	for length / curr > Unit {
		curr +=1
	}
	return curr
}

func TestSplit(t *testing.T) {
	area := fetcher.GetOblong(
		function.Vector{
			0, 0, 0	,
		},function.Vector{
			201, 200, 201,
		},
	)
	fmt.Println(area, len(area))
	fmt.Println(SplitArea(1111))
}