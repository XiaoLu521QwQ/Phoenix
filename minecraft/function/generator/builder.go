package generator

import (
	"errors"
	"fmt"
	"phoenix/minecraft/function"
	"phoenix/minecraft/ligo"
)

func PluginInit(vm *ligo.VM){
	vm.Funcs["circle"] = Circle

}

func getFloat(vars ...ligo.Variable) ([]float64, error) {
	var res = []float64{}
	for k, v := range vars {
		if v.Type == ligo.TypeFloat {
			res = append(res, v.Value.(float64))
		} else if v.Type == ligo.TypeInt {
			res = append(res, float64(v.Value.(int64)))
		} else {
			return nil, errors.New(fmt.Sprintf("getFloat: expected a Int or float type, got %v at %v", v.Type, k))
		}
	}
	return res, nil
}

// Circle : (circle radius inner-radius height facing)
func Circle(_vm *ligo.VM, a ...ligo.Variable) ligo.Variable {
	vars, err := getFloat(a[:3]...)
	if err != nil {
		return ligo.Variable{
			Type:  ligo.TypeErr,
			Value: err,
		}
	}
	radius := vars[0]
	inner := vars[1]
	height := vars[2]
	facing := a[3].Value.(string)
	var vec []function.Vector
	switch facing {
	case "x":
		for h := 0.0; h < height ; h += 1.0 {
			for x := -radius ; x < radius ; x++ {
				for y := -radius ; y < radius ; y++ {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{h, x, y})
					}
				}
			}
		}
	case "y":
		for h := 0.0; h < height ; h += 1.0 {
			for x := -radius ; x < radius ; x++ {
				for y := -radius ; y < radius ; y++ {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{x, h, y})
					}
				}
			}
		}
	case "z":
		for h := 0.0; h < height ; h += 1.0 {
			for x := -radius ; x < radius ; x++ {
				for y := -radius ; y < radius ; y++ {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{h, x, y})
					}
				}
			}
		}
	}

	return ligo.Variable{
		Type:  ligo.TypeArray,
		Value: vec,
	}
}

