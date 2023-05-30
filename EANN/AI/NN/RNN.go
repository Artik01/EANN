package NN

import (
	"fmt"
)

type (
	RNN struct {
		Weights []Matrix
		Save Vector
		In int
		Out int
	}
)

var RNNtempl *RNN = new(RNN)

func (nn RNN) GetType() string {
	return "RNN"
}

func (nn RNN) New(params ...int) NN {
	var newnn RNN
	rec,_:=NewMatrix(params[1]+1,params[1])
	newnn.Weights = append(newnn.Weights, rec)
	for i:=0; i<len(params)-1;i++ {
		m,_:=NewMatrix(params[i]+1,params[i+1])
		newnn.Weights = append(newnn.Weights, m)
	}
	newnn.In = params[0]
	newnn.Out = params[len(params)-1]
	newnn.Save,_=NewVector(params[1])
	return &newnn
}

func (nn *RNN) Calculate(in Vector) (Vector, error) {
	if len(in)!=nn.In {return nil, fmt.Errorf("Incorrect input size")}
	v := make(Vector, len(in))
	copy(v, in)
	save := make(Vector, len(nn.Save))
	copy(save, nn.Save)
	save.AdjustForBias()
	a,err:=nn.Weights[0].MultBy(save)
	if err!=nil {return nil, fmt.Errorf("Mult!")}
	v.AdjustForBias()
	b, err:=nn.Weights[1].MultBy(v)
	if err!=nil {return nil, fmt.Errorf("Mult!(2)")}
	v,err=a.PointwiseSum(b)
	if err!=nil {return nil, fmt.Errorf("Sum!")}
	v.Activate(Sigmoid)//or tanh
	for i:=2; i<len(nn.Weights);i++ {
		m := nn.Weights[i]
		v.AdjustForBias()
		v,_=m.MultBy(v)
		v.Activate(Sigmoid)//or tanh
	}
	copy(nn.Save,v)
	return v, nil
}

func (nn *RNN) Clear() {
	(*nn).Weights = nil
	(*nn).Save,_=NewVector(len((*nn).Save))
}

func (nn *RNN) GenFromIndividual(i Individual) error {
	if len(i) < 1 {return fmt.Errorf("Incorrect input size")}
	nn.Clear()
	(*nn).Weights = i
	return nil
}

func (nn RNN) MakeIndividual() Individual {
	return Individual(nn.Weights)
}
