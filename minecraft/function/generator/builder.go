package generator

import (
	"errors"
	"fmt"
	"phoenix/minecraft/function"
	"phoenix/minecraft/ligo"
)

func PluginInit(vm *ligo.VM){
	vm.Funcs["circle"] = Circle
	vm.Funcs["sphere"] = Sphere
	vm.Funcs["ellipse"] = Ellipse
	vm.Funcs["comp"] = Composition
}

type VMFunc func(vm *ligo.VM, v...ligo.Variable) ligo.Variable

// Composition : (composition function list)
func Composition(vm *ligo.VM, a ...ligo.Variable) ligo.Variable{
	var res []ligo.Variable
	if a[0].Type == ligo.TypeIFunc && a[1].Type == ligo.TypeArray {
		fn := a[0].Value.(VMFunc)
		for _,v := range a[1].Value.([]ligo.Variable) {
			res = append(res, fn(vm, v))
		}
	}
	return ligo.Variable{
		Type:  ligo.TypeArray,
		Value: res,
	}
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
func Circle(vm *ligo.VM, a ...ligo.Variable) ligo.Variable {
	vars, err := getFloat(a[:3]...)
	if err != nil {
		return vm.Throw(fmt.Sprintf("%s", err))
	}
	radius := vars[0]
	inner := vars[1]
	height := vars[2]
	facing := a[3].Value.(string)
	var vec []function.Vector
	switch facing {
	case "x":
		for h := 0.0; h <= height ; h += 1.0 {
			for x := -radius ; x <= radius ; x += 1.0 {
				for y := -radius ; y < radius ; y += 1.0 {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{h, x, y})
					}
				}
			}
		}
	case "y":
		for h := 0.0; h <= height ; h += 1.0 {
			for x := -radius ; x <= radius ; x += 1.0 {
				for y := -radius ; y <= radius ; y += 1.0 {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{x, h, y})
					}
				}
			}
		}
	case "z":
		for h := 0.0; h <= height ; h += 1.0 {
			for x := -radius ; x <= radius ; x += 1.0 {
				for y := -radius ; y <= radius ; y += 1.0 {
					if radius * radius > x * x + y * y && x * x + y * y >= (radius - inner) * (radius - inner) {
						vec = append(vec, []float64{h, x, y})
					}
				}
			}
		}
		default:
			return vm.Throw(fmt.Sprintf("circle: "))
	}

	return ligo.Variable{
		Type:  ligo.TypeArray,
		Value: vec,
	}
}

// Sphere : (sphere radius inner-radius
func Sphere(vm *ligo.VM, a ...ligo.Variable) ligo.Variable {
	vars, err := getFloat(a...)
	if err != nil {
		return vm.Throw(fmt.Sprintf("%s", err))
	}

	r := vars[0]
	ir := vars[1]
	if r < ir {
		return vm.Throw(fmt.Sprintf("sphere: Inner radius (%v) is larger than radius (%v)", ir, r))
	}
	var vec []function.Vector
	for x := -r ; x < r ; x++ {
		for y := -r ; y < r ; y++ {
			for z := -r ; z < r ; z++ {
				if r*r>=x*x+y*y+z*z&&x*x+y*y+z*z>=ir*ir{
					vec = append(vec, []float64{x, y, z})
				}
			}
		}
	}
	return ligo.Variable{
		Type:  ligo.TypeArray,
		Value: vec,
	}
}

// Ellipse : (ellipse width length height facing
func Ellipse(vm *ligo.VM, a ...ligo.Variable) ligo.Variable {
	vars, err := getFloat(a[:3]...)
	if err != nil {
		return vm.Throw(fmt.Sprintf("%s", err))
	}
	width := vars[0]
	length := vars[1]
	height := vars[2]
	facing := a[3].Value.(string)
	var vec []function.Vector
	switch facing {
	case "x":
		for h := 0.0; h <= height ; h += 1.0 {
			for i := -length ; i <= length ; i += 1.0 {
				for j := -width; j <= width ; j += 1.0 {
					if (i * i * 1.0) / (length * length) + (j * j * 1.0) / (width * width) < 1 {
						vec = append(vec, []float64{h, i, j})
					}
				}
			}
		}
	case "y":
		for h := 0.0; h <= height ; h += 1.0 {
			for i := -length ; i <= length ; i += 1.0 {
				for j := -width; j <= width ; j += 1.0 {
					if (i * i * 1.0) / (length * length) + (j * j * 1.0) / (width * width) < 1 {
						vec = append(vec, []float64{i, j, h})
					}
				}
			}
		}
	case "z":
		for h := 0.0; h <= height ; h += 1.0 {
			for i := -length ; i <= length ; i += 1.0 {
				for j := -width; j <= width ; j += 1.0 {
					if (i * i * 1.0) / (length * length) + (j * j * 1.0) / (width * width) < 1 {
						vec = append(vec, []float64{i, h, j})
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

