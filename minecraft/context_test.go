package minecraft_test

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"testing"
)

func TestMat(t *testing.T) {
	v := []float64{1,2,3,4,5,6,7,8,9}
	A := mat.NewDense(3, 3, v)
	B := mat.NewDense(3, 3, []float64{2,2,2,2,2,2,2,2,2})
	var C *mat.Dense
	C.Product(A, B)
	fmt.Println(mat.Formatted(C, mat.Prefix(" "), mat.Squeeze()))
	fmt.Println(v)
}