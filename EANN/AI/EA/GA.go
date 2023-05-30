package EA

import (
	"math/rand"
)

type (
	GA struct {
		Popul Population
		fitF FitFunc
	}
	Population []GAIndividual
	GAIndividual struct{
		Ind Individual
		Fit ContextFit
	}
)

var GAtempl *GA = new(GA)
const GAmutProb = 0.05

func (alg GA) GetType() string {
	return "GA"
}

func (alg GA) GetBest() (ContextFit, Individual) {
	return alg.Popul[0].Fit, alg.Popul[0].Ind
}

func (alg *GA) ResetTop() {
	return
}

func (ind GAIndividual) Copy() GAIndividual {
	var newi GAIndividual
	b:= ind.Ind
	b.Copy(ind.Ind)
	newi.Ind=b
	newi.Fit=ind.Fit.Copy()
	return newi
}

func (alg GA) GetCurrFits() []ContextFit {
	var res []ContextFit
	for _, ind := range alg.Popul {
		res = append(res, ind.Fit)
	}
	return res
}

func (alg *GA) StartUp(ind Individual, ff FitFunc, size int) {
	alg.fitF=ff
	Popul:=GenPopulation(ind, size)
	alg.Popul = make(Population, len(Popul))
	for i := range Popul {
		alg.Popul[i]=GAIndividual{Popul[i],MInf}
	}
	for i := range alg.Popul {
		res:=alg.fitF(alg.Popul[i].Ind)
		alg.Popul[i].Fit=res
	}
}

func (alg *GA) Iterate() {
	n:=len(alg.Popul)
	parents:=alg.Popul.Select(n)
	for i:=0; i < len(parents)-1; i+=2 {
		new1, new2:=parents[i].Ind.Crossover(parents[i+1].Ind)
		new1.Mutate(GAmutProb, SSLowerB, SSUpperB)
		new2.Mutate(GAmutProb, SSLowerB, SSUpperB)
		alg.Popul = append(alg.Popul,GAIndividual{new1,MInf})
		alg.Popul = append(alg.Popul,GAIndividual{new2,MInf})
	}
	for i, ind := range alg.Popul {
		fit:=alg.fitF(ind.Ind)
		alg.Popul[i].Fit=fit
	}
	
	Elite := alg.Popul[0]
	for i := range alg.Popul {
		if alg.Popul[i].Fit.GetFit() > Elite.Fit.GetFit() {
			Elite = alg.Popul[i]
		}
	}
	alg.Popul = Population{Elite}
	
	alg.Popul=append(alg.Popul, alg.Popul.Select(n-1)...)
}

func (alg GA) New() EA {
	new := GA{Population{}, nil}
	return &new
}

func (p Population) Select(n int) Population {
	var new Population
	var sum float64
	for _, ind := range p {
		sum += ind.Fit.GetFit()
	}
	for i := 0; i<n; i++ {
		var k int
		pvar := rand.Float64()
		var psum float64
		for j, ind := range p {
			psum += ind.Fit.GetFit()/sum
			if psum >= pvar {
				k = j
				break
			}
		}
		new = append(new, p[k].Copy())
	}
	return new
}
