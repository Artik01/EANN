package NN

import (
	"fmt"
	"math"
	"math/rand"
)

type (
	NN interface {
		Calculate(Vector) (Vector, error)
		Clear()
		GenFromIndividual(i Individual) error
		New(...int) NN
		MakeIndividual() Individual
		GetType() string
	}
	Matrix [][]float64
	Vector []float64
	
	ActivationFunc func(float64) float64
)

var Sigmoid ActivationFunc = func(x float64) float64 {return 1.0/(1+math.Exp(-x))}
var Tanh ActivationFunc = func(x float64) float64 {return math.Tanh(x)}

func NewVector(l int) (Vector,error) {
	if l<=0 {return nil, fmt.Errorf("Vector length cannot be <=0")}
	return make(Vector, l), nil
}

func NewMatrix(w, h int) (Matrix,error) {
	if w<=0 || h<=0 {return nil, fmt.Errorf("Matrix size cannot be <=0")}
	var m Matrix = make(Matrix, h)
	for i:=range m {
		m[i]=make([]float64,w)
	}
	return m, nil
}

func (m Matrix) MultBy(v Vector) (Vector,error) {
	width := len(m[0])
	height := len(m)
	if width!=len(v) {return nil, fmt.Errorf("Cannot mult matrix and vector")}
	res, err := NewVector(height)
	if err !=nil {
		return nil, err
	}
	
	for i := range res {
		var sum float64 = 0
		for j, a := range m[i] {
			sum += a*v[j]
		}
		res[i]=sum
	}
	return res, nil
}

func (m Matrix) Sum(m2 Matrix) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]+m2[i][j]
		}
	}
	return newM
}

func (m Matrix) Add(p float64) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]+p
		}
	}
	return newM
}

func (m Matrix) Sub(m2 Matrix) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]-m2[i][j]
		}
	}
	return newM
}

func (m Matrix) Mult(p float64) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]*p
		}
	}
	return newM
}

func (m Matrix) Div(p float64) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]/p
		}
	}
	return newM
}

func (m Matrix) PointwiseMult(m2 Matrix) Matrix {
	newM, _ := NewMatrix(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			newM[i][j]=m[i][j]*m2[i][j]
		}
	}
	return newM
}

func (m Matrix) Random(low, up float64) Matrix {
	new := make(Matrix, len(m))
	
	for i:=range new {
		new[i]=make([]float64,len(m[i]))
		for j := range new[i] {
			new[i][j]= rand.Float64()*(up-low)+low
		}
	}
	return new
}

func (m Matrix) Len() int {
	return len(m)*len(m[0])
}

func (m Matrix) Get(n int) float64 {
	w := len(m[0])
	i := n/w
	j := n%w
	return m[i][j]
}

func (m *Matrix) Set(n int, p float64)  {
	w := len((*m)[0])
	i := n/w
	j := n%w
	(*m)[i][j]=p
}

func (m *Matrix) Copy(m2 Matrix) {
	*m=make(Matrix, len(m2))
	for i:=range *m {
		(*m)[i]=make([]float64, len(m2[i]))
		for j:=range (*m)[i] {
			(*m)[i][j]=m2[i][j]
		}
	}
}

func (v *Vector) Activate(f ActivationFunc) {
	for i := range *v {
		(*v)[i]=f((*v)[i])
	}
}

func (v *Vector) AdjustForBias() {
	*v = append(Vector{1}, (*v)...)
}

func (h Vector) Append(x Vector) Vector {
	return append(h, x...)
}

func (a Vector) PointwiseMult(b Vector) (Vector,error) {
	if len(a)!=len(b) {return nil, fmt.Errorf("Cannot mult vectors")}
	res := make(Vector, len(a))
	for i:=range a {
		res[i]=a[i]*b[i]
	}
	return res,nil
}

func (a Vector) PointwiseSum(b Vector) (Vector,error) {
	if len(a)!=len(b) {return nil, fmt.Errorf("Cannot sum vectors")}
	res := make(Vector, len(a))
	for i:=range a {
		res[i]=a[i]+b[i]
	}
	return res,nil
}

func New[NNT NN](templ NNT, sizes ...int) (NNT,error) {
	if len(sizes)<2 {return templ, fmt.Errorf("Cannot create NN")}
	for _, s := range sizes {
		if s<=0 {return templ, fmt.Errorf("Cannot create NN")}
	}
	return templ.New(sizes...).(NNT), nil
}
