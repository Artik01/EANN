package EA

import (
	"math/rand"
)

type (
	ABC struct {
		EmployeesBees Employees
		OnlookersBees Onlookers
		fitF FitFunc
		BestPos Individual
		BestFit ContextFit
	}
	Employees []Employee
	Onlookers []int
	Employee struct{
		Solution Individual
		Fit ContextFit
		Unimproved int
		Improved bool
	}
)

var ABCtempl *ABC = new(ABC)
const Limit int = 15

func (alg ABC) GetType() string {
	return "ABC"
}

func (alg ABC) GetBest() (ContextFit, Individual) {
	return alg.BestFit, alg.BestPos
}

func (alg *ABC) ResetTop() {
	alg.BestFit = alg.BestFit.Reset()
}

func (alg ABC) GetCurrFits() []ContextFit {
	var res []ContextFit
	for _, ind := range alg.EmployeesBees {
		res = append(res, ind.Fit)
	}
	return res
}

func (alg *ABC) StartUp(ind Individual, ff FitFunc, size int) {
	alg.fitF=ff
	Popul:=GenPopulation(ind, size/2)
	alg.EmployeesBees = make(Employees, len(Popul))
	for i := range Popul {
		alg.EmployeesBees[i]=Employee{Popul[i], MInf, 0, false}
	}
	alg.BestPos = alg.EmployeesBees[0].Solution
	for i := range alg.EmployeesBees {
		res:=alg.fitF(alg.EmployeesBees[i].Solution)
		alg.EmployeesBees[i].Fit=res
		if res.GetFit() >= alg.BestFit.GetFit() {
			alg.BestFit = res
			alg.BestPos = alg.EmployeesBees[i].Solution
		}
	}
	alg.OnlookersBees = make(Onlookers, size-size/2)
	for i := range alg.OnlookersBees {
		alg.OnlookersBees[i]=0
	}
}

func (alg *ABC) Iterate() {
	for i, emp := range alg.EmployeesBees {
		alg.EmployeesBees[i].Improved = false
		if emp.Unimproved >= Limit {
			alg.EmployeesBees[i].Improved = true
			alg.EmployeesBees[i].Unimproved = 0
			alg.EmployeesBees[i].Solution=emp.Solution.Random(SSLowerB, SSUpperB)
			fit:=alg.fitF(alg.EmployeesBees[i].Solution)
			alg.EmployeesBees[i].Fit=fit
		} else {
			soli:=emp.Solution
			j:=rand.Intn(len(alg.EmployeesBees))
			for i==j {
				j=rand.Intn(len(alg.EmployeesBees))
			}
			solj:=alg.EmployeesBees[j].Solution
			k:=rand.Intn(soli.Len())
			
			xik:=soli.Get(k)
			v:=xik+(rand.Float64()*2-1)*(xik-solj.Get(k))
			newSol := soli
			newSol.Copy(soli)
			newSol.Set(k, v)
			fit:=alg.fitF(newSol)
			if fit.GetFit() > emp.Fit.GetFit() {
				alg.EmployeesBees[i].Solution = newSol
				alg.EmployeesBees[i].Fit = fit
				alg.EmployeesBees[i].Improved = true
				alg.EmployeesBees[i].Unimproved = 0
				if fit.GetFit() >= alg.BestFit.GetFit() {
					alg.BestFit = fit
					alg.BestPos = newSol
				}
			}
		}
	}
	for i := range alg.OnlookersBees {
		var sum float64
		for _, emp := range alg.EmployeesBees {
			sum += emp.Fit.GetFit()
		}
		var k int
		p := rand.Float64()
		var psum float64
		for i, emp := range alg.EmployeesBees {
			psum += emp.Fit.GetFit()/sum
			if psum > p {
				k = i
				break
			}
		}
		alg.OnlookersBees[i]=k
	}
	for n := range alg.OnlookersBees {
		i:=alg.OnlookersBees[n]
		emp := alg.EmployeesBees[i]
		soli:=emp.Solution
		j:=rand.Intn(len(alg.EmployeesBees))
		for i==j {
			j=rand.Intn(len(alg.EmployeesBees))
		}
		solj:=alg.EmployeesBees[j].Solution
		k:=rand.Intn(soli.Len())
		
		xik:=soli.Get(k)
		v:=xik+(rand.Float64()*2-1)*(xik-solj.Get(k))
		newSol := soli
		newSol.Copy(soli)
		newSol.Set(k, v)
		fit:=alg.fitF(newSol)
		if fit.GetFit() > emp.Fit.GetFit() {
			alg.EmployeesBees[i].Solution = newSol
			alg.EmployeesBees[i].Fit = fit
			alg.EmployeesBees[i].Improved = true
			alg.EmployeesBees[i].Unimproved = 0
			if fit.GetFit() >= alg.BestFit.GetFit() {
				alg.BestFit = fit
				alg.BestPos = newSol
			}
		}
	}
	for i, emp := range alg.EmployeesBees {
		if !emp.Improved {
			alg.EmployeesBees[i].Unimproved++
		}
	}
}

func (alg ABC) New() EA {
	new := ABC{Employees{}, Onlookers{}, nil, nil, MInf}
	return &new
}
