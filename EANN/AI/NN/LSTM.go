package NN

import (
	"fmt"
)

type (
	LSTM struct {
		ForgetGM Matrix
		InputGM Matrix
		NewMNetM Matrix
		OutputGM Matrix
		
		CellState Vector
		HiddenState Vector
		
		In int
		Out int
	}
)

var LSTMtempl *LSTM = new(LSTM)

func (nn LSTM) GetType() string {
	return "LSTM"
}

func (nn LSTM) New(params ...int) NN {
	var newnn LSTM
	
	newnn.In = params[0]
	newnn.Out = params[1]
	
	newnn.ForgetGM,_ = NewMatrix(newnn.In+newnn.Out+1,newnn.Out)
	newnn.InputGM,_ = NewMatrix(newnn.In+newnn.Out+1,newnn.Out)
	newnn.NewMNetM,_ = NewMatrix(newnn.In+newnn.Out+1,newnn.Out)
	newnn.OutputGM,_ = NewMatrix(newnn.In+newnn.Out+1,newnn.Out)
	newnn.CellState,_ = NewVector(newnn.Out)
	newnn.HiddenState,_ = NewVector(newnn.Out)
	
	return &newnn
}

func (nn *LSTM) Calculate(in Vector) (Vector, error) {
	if len(in)!=nn.In {return nil, fmt.Errorf("Incorrect input size")}
	
	x := make(Vector, len(in))
	copy(x, in)
	h := make(Vector, len(nn.HiddenState))
	copy(h, nn.HiddenState)
	
	buffV := h.Append(x)
	buffV.AdjustForBias()
	
	f,err := nn.ForgetGM.MultBy(buffV)
	if err!=nil {return nil, fmt.Errorf("Mult 1")}
	f.Activate(Sigmoid)
	i,err := nn.InputGM.MultBy(buffV)
	if err!=nil {return nil, fmt.Errorf("Mult 2")}
	i.Activate(Sigmoid)
	Cbuff,err := nn.NewMNetM.MultBy(buffV)
	if err!=nil {return nil, fmt.Errorf("Mult 3")}
	Cbuff.Activate(Tanh)
	o,err := nn.OutputGM.MultBy(buffV)
	if err!=nil {return nil, fmt.Errorf("Mult 4")}
	o.Activate(Sigmoid)
	
	C := make(Vector, len(nn.CellState))
	copy(C, nn.CellState)
	
	C,err = C.PointwiseMult(f)
	if err!=nil {return nil, fmt.Errorf("Mult 11")}
	
	Cbuff,err = Cbuff.PointwiseMult(i)
	if err!=nil {return nil, fmt.Errorf("Mult 12")}
	
	C,err = C.PointwiseSum(Cbuff)
	if err!=nil {return nil, fmt.Errorf("Sum")}
	
	copy(nn.CellState,C)
	C.Activate(Tanh)
	
	h,err = C.PointwiseMult(o)
	if err!=nil {return nil, fmt.Errorf("Mult 21")}
	
	copy(nn.HiddenState,h)
	
	return h, nil
}

func (nn *LSTM) Clear() {
	(*nn).ForgetGM,(*nn).InputGM,(*nn).NewMNetM,(*nn).OutputGM = nil,nil,nil,nil
	(*nn).CellState, _ = NewVector(nn.Out)
	(*nn).HiddenState, _ = NewVector(nn.Out)
}

func (nn *LSTM) GenFromIndividual(i Individual) error {
	if len(i) < 4 {return fmt.Errorf("Incorrect input size")}
	nn.Clear()
	(*nn).ForgetGM,(*nn).InputGM,(*nn).NewMNetM,(*nn).OutputGM = i[0],i[1],i[2],i[3]
	return nil
}

func (nn LSTM) MakeIndividual() Individual {
	return Individual{nn.ForgetGM,nn.InputGM,nn.NewMNetM,nn.OutputGM}
}
