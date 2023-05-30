package EA

import (
	"math"
)

type (
	PSO struct {
		Swarm Particles
		Best Individual
		BestFit ContextFit
		fitF FitFunc
	}
	Particles []Particle
	Particle struct {
		Passanger Individual
		Velocity Individual
		Best Individual
		BestFit ContextFit
		CurrFit ContextFit
	}
)

var PSOtempl *PSO = new(PSO)
const w, φp, φg float64 = 0.6, 0.4, 0.2

func (alg PSO) GetType() string {
	return "PSO"
}

func (alg PSO) GetBest() (ContextFit, Individual) {
	return alg.BestFit, alg.Best
}

func (alg *PSO) ResetTop() {
	fit := alg.fitF(alg.Best)
	alg.BestFit = fit
}

func (alg PSO) GetCurrFits() []ContextFit {
	var res []ContextFit
	for _, ind := range alg.Swarm {
		res = append(res, ind.CurrFit)
	}
	return res
}

func (alg *PSO) StartUp(ind Individual, ff FitFunc, size int) {
	alg.fitF=ff
	Popul:=GenPopulation(ind, size)
	alg.Swarm = make(Particles, len(Popul))
	for i := range Popul {
		vel:=ind.Random(-math.Abs(SSUpperB-SSLowerB), math.Abs(SSUpperB-SSLowerB))
		alg.Swarm[i]=Particle{Popul[i],vel,Popul[i],MInf, MInf}
		alg.Swarm[i].Best.Copy(alg.Swarm[i].Passanger)
	}
	alg.Best=alg.Swarm[0].Passanger
	alg.Best.Copy(alg.Swarm[0].Passanger)
	for i := range alg.Swarm {
		res:=alg.fitF(alg.Swarm[i].Passanger)
		alg.Swarm[i].BestFit=res
		alg.Swarm[i].CurrFit=res
		if alg.BestFit.GetFit() < res.GetFit() {
			alg.BestFit = res
			alg.Best.Copy(alg.Swarm[i].Passanger)
		}
	}
}

func (alg *PSO) Iterate() {
	var newSwarm Particles
	for i := range alg.Swarm {
		part:=alg.Swarm[i]
		rp:=part.Passanger.Random(0,1)
		rg:=part.Passanger.Random(0,1)
		newV:=part.Velocity.Mult(w).Add(part.Best.Sub(part.Passanger).PointwiseMult(rp).Mult(φp)).Add(alg.Best.Sub(part.Passanger).PointwiseMult(rg).Mult(φg))
		part.Velocity = newV
		part.Passanger = part.Passanger.Add(newV)
		fit := alg.fitF(part.Passanger)
		part.CurrFit = fit
		if fit.GetFit() > part.BestFit.GetFit() {
			part.BestFit = fit
			part.Best.Copy(part.Passanger)
			if fit.GetFit() > alg.BestFit.GetFit() {
				alg.BestFit = fit
				alg.Best.Copy(part.Passanger)
			}
		}
		newSwarm=append(newSwarm,part)
	}
	alg.Swarm=newSwarm
}

func (alg PSO) New() EA {
	new := PSO{Particles{}, nil, MInf, nil}
	return &new
}
