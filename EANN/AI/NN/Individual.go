package NN

import (
	"AI/EA"
	"math/rand"
	"math"
)

type(
	Individual []Matrix
)

func (p1 Individual) Crossover(p2t EA.Individual) (EA.Individual, EA.Individual) {
	p2:=*(p2t.(*Individual))
	ch1, ch2 := *(p1.Random(0,1).(*Individual)), *(p2.Random(0,1).(*Individual))
	p := rand.Float64()
	for i := 0; i<len(p1); i++ {
		for j := 0; j<len(p1[i]); j++ {
			for k := 0; k<len(p1[i][j]); k++ {
				ch1[i][j][k]=p1[i][j][k]*p+p2[i][j][k]*(1-p)
				ch2[i][j][k]=p2[i][j][k]*p+p1[i][j][k]*(1-p)
			}
		}
	}
	return &ch1, &ch2
}

func (ind *Individual) Mutate(prob float64, low, up float64) {
	if prob < 0 || prob > 1 {return}
	if rand.Float64() < prob {
		i := rand.Intn(len(*ind))
		(*ind)[i] = (*ind)[i].Random(low, up)
	}
}

func (ind Individual) Add(i2 EA.Individual) EA.Individual {
	res := make(Individual, len(ind))
	ind2 := *(i2.(*Individual))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].Sum(ind2[i])
	}
	return &res
}

func (ind Individual) GenAdd(p float64) EA.Individual {
	res := make(Individual, len(ind))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].Add(p)
	}
	return &res
}

func (ind Individual) Sub(i2 EA.Individual) EA.Individual {
	res := make(Individual, len(ind))
	ind2 := *(i2.(*Individual))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].Sub(ind2[i])
	}
	return &res
}

func (ind Individual) Mult(p float64) EA.Individual {
	res := make(Individual, len(ind))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].Mult(p)
	}
	return &res
}

func (ind Individual) Div(p float64) EA.Individual {
	res := make(Individual, len(ind))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].Div(p)
	}
	return &res
}

func (ind Individual) PointwiseMult(i2 EA.Individual) EA.Individual {
	res := make(Individual, len(ind))
	ind2 := *(i2.(*Individual))
	for i := 0; i < len(ind); i++ {
		res[i] = ind[i].PointwiseMult(ind2[i])
	}
	return &res
}

func (templ Individual) Random(low, up float64) EA.Individual {
	res := make(Individual, len(templ))
	for i, m := range templ {
		res[i] = m.Random(low, up)
	}
	return &res
}

func (ind Individual) Get(n int) float64 {
	var l int
	var oldl int
	for i := range ind {
		oldl=l
		l+=ind[i].Len()
		if l>n {
			return ind[i].Get(n-oldl)
		}
	}
	return 0
}

func (ind *Individual) Set(n int, p float64) {
	var l int
	var oldl int
	for i := range *ind {
		oldl=l
		l+=(*ind)[i].Len()
		if l>n {
			(*ind)[i].Set(n-oldl, p)
			return
		}
	}
}

func (ind Individual) Len() int {
	var l int
	for i := range ind {
		l+=ind[i].Len()
	}
	return l
}

func (ind *Individual) Copy(i2 EA.Individual) {
	ind2 := *(i2.(*Individual))
	*ind = make(Individual, len(ind2))
	for i := 0; i < len(*ind); i++ {
		(*ind)[i].Copy(ind2[i])
	}
}

func (ind Individual) Distance(ind2 EA.Individual) float64 {
	difft := ind.Sub(ind2)
	var sum float64
	diff := *(difft.(*Individual))
	for i := range diff {
		for j := range diff[i] {
			for k := range diff[i][j] {
				a:=diff[i][j][k]
				sum+=a*a
			}
		}
	}
	return math.Sqrt(sum)
}
