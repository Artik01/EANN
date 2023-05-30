package NN

import (
	"fmt"
)

type (
	FNN struct {
		Weights []Matrix
		In int
		Out int
	}
)

var FNNtempl *FNN = new(FNN)

func (nn FNN) GetType() string {
	return "FNN"
}

func (nn FNN) New(params ...int) NN {
	var newnn FNN
	for i:=0; i<len(params)-1;i++ {
		m,_:=NewMatrix(params[i]+1,params[i+1])
		newnn.Weights = append(newnn.Weights, m)
	}
	newnn.In = params[0]
	newnn.Out = params[len(params)-1]
	return &newnn
}

func (nn *FNN) Calculate(in Vector) (Vector, error) {
	if len(in)!=nn.In {return nil, fmt.Errorf("Incorrect input size")}
	v := make(Vector, len(in))
	copy(v, in)
	for _, m := range nn.Weights {
		v.AdjustForBias()
		v,_=m.MultBy(v)
		v.Activate(Sigmoid)//or tanh
	}
	return v, nil
}

func (nn *FNN) Clear() {
	(*nn).Weights = nil
}

func (nn *FNN) GenFromIndividual(i Individual) error {
	if len(i) < 1 {return fmt.Errorf("Incorrect input size")}
	nn.Clear()
	(*nn).Weights = i
	return nil
}

func (nn FNN) MakeIndividual() Individual {
	return Individual(nn.Weights)
}
