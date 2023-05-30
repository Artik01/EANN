package EA

import (
	"math"
	"math/rand"
)

type (
	SSO struct {
		Males Spiders
		Females Spiders
		fitF FitFunc
		worst, best ContextFit
		besti SpiderInfo
		BestPos Individual
		BestFit ContextFit
	}
	Spiders []Spider
	Spider struct {
		Passanger Individual
		Fit ContextFit
		Weight float64
	}
	SpiderInfo struct {
		t gender
		i int
	}
	gender int
	FilterF func(s Spider) bool
)

var SSOtempl *SSO = new(SSO)
const (
	male gender = 2<<iota+1
	female
)
const PF float64 = 0.6
const radius float64 = (SSUpperB - SSLowerB)/2
const sChance float64 = 0.25
const SSOmutProb = 0.05

func (alg SSO) GetType() string {
	return "SSO"
}

func (alg SSO) GetBest() (ContextFit, Individual) {
	return alg.BestFit, alg.BestPos
}

func (alg *SSO) ResetTop() {
	alg.BestFit = alg.BestFit.Reset()
}

func (alg SSO) GetCurrFits() []ContextFit {
	var res []ContextFit
	for _, ind := range alg.Males {
		res = append(res, ind.Fit)
	}
	for _, ind := range alg.Females {
		res = append(res, ind.Fit)
	}
	return res
}

func (alg *SSO) StartUp(ind Individual, ff FitFunc, size int) {
	alg.fitF=ff
	Nf:=int(math.Floor(float64(size)*(0.9-rand.Float64()*0.25)))
	Nm:=size-Nf
	Popul:=GenPopulation(ind, Nm)
	alg.Males = make(Spiders, len(Popul))
	for i := range Popul {
		alg.Males[i]=Spider{Popul[i], MInf, math.Inf(-1)}
	}
	for i := range alg.Males {
		res:=alg.fitF(alg.Males[i].Passanger)
		alg.Males[i].Fit=res
	}
	alg.BestPos=alg.Males[0].Passanger
	alg.BestPos.Copy(alg.Males[0].Passanger)
	Popul=GenPopulation(ind, Nf)
	alg.Females = make(Spiders, len(Popul))
	for i := range Popul {
		alg.Females[i]=Spider{Popul[i], MInf, math.Inf(-1)}
	}
	for i := range alg.Females {
		res:=alg.fitF(alg.Females[i].Passanger)
		alg.Females[i].Fit=res
	}
	
	var worst, best ContextFit = alg.Males[0].Fit, alg.Males[0].Fit
	var besti SpiderInfo = SpiderInfo{male, 0}
	for i := range alg.Males {
		if worst.GetFit() > alg.Males[i].Fit.GetFit() {
			worst = alg.Males[i].Fit
		}
		if best.GetFit() < alg.Males[i].Fit.GetFit() {
			best = alg.Males[i].Fit
			besti = SpiderInfo{male, i}
			if alg.BestFit.GetFit() <= best.GetFit() {
				alg.BestFit = best
				alg.BestPos.Copy(alg.Males[i].Passanger)
			}
		}
	}
	for i := range alg.Females {
		if worst.GetFit() > alg.Females[i].Fit.GetFit() {
			worst = alg.Females[i].Fit
		}
		if best.GetFit() < alg.Females[i].Fit.GetFit() {
			best = alg.Females[i].Fit
			besti = SpiderInfo{female, i}
			if alg.BestFit.GetFit() <= best.GetFit() {
				alg.BestFit = best
				alg.BestPos.Copy(alg.Females[i].Passanger)
			}
		}
	}
	alg.worst=worst
	alg.best=best
	alg.besti=besti
	
	for i := range alg.Males {
		alg.Males[i].Weight = (alg.Males[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
	for i := range alg.Females {
		alg.Females[i].Weight = (alg.Females[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
}

func (alg *SSO) Iterate() {
	fcopy := make(Spiders, len(alg.Females))
	mcopy := make(Spiders, len(alg.Males))
	copy(fcopy,alg.Females)
	copy(mcopy,alg.Males)
	
	for i := range fcopy {
		if alg.Females[i].Fit.GetFit()>=alg.best.GetFit() {
			continue
		}
		alpha,beta,delta,r:=rand.Float64(),rand.Float64(),rand.Float64(),rand.Float64()
		sc:=alg.Females[i].FindBetterClosest(append(alg.Females, alg.Males...))
		Vibc:=alg.Females[i].VibWith(sc)
		sb:=alg.Find(alg.besti)
		Vibb:=alg.Females[i].VibWith(sb)
		p := rand.Float64()
		pos:=alg.Females[i].Passanger
		if p < PF {
			fcopy[i].Passanger=pos.Add(sc.Passanger.Sub(pos).Mult(alpha*Vibc)).Add(sb.Passanger.Sub(pos).Mult(beta*Vibb)).GenAdd(delta*(r-1/2))
		} else {
			fcopy[i].Passanger=pos.Sub(sc.Passanger.Sub(pos).Mult(alpha*Vibc)).Sub(sb.Passanger.Sub(pos).Mult(beta*Vibb)).GenAdd(delta*(r-1/2))
		}
	}
	
	med:=alg.Males.Median().Weight
	mean:= alg.Males.Mean()
	for i := range mcopy {
		alpha,delta,r:=rand.Float64(),rand.Float64(),rand.Float64()
		sf:=alg.Males[i].FindClosest(alg.Females)
		Vibf:=alg.Males[i].VibWith(sf)
		IsDominant := alg.Males[i].Weight>med
		pos:=alg.Males[i].Passanger
		if IsDominant {
			mcopy[i].Passanger=pos.Add(sf.Passanger.Sub(pos).Mult(Vibf*alpha)).GenAdd(delta*(r-1/2))
		} else {
			mcopy[i].Passanger=pos.Add(mean.Sub(pos).Mult(alpha))
		}
	}
	
	for i := range fcopy {
		res:=alg.fitF(fcopy[i].Passanger)
		fcopy[i].Fit=res
	}
	for i := range mcopy {
		res:=alg.fitF(mcopy[i].Passanger)
		mcopy[i].Fit=res
	}
	
	var worst, best ContextFit = mcopy[0].Fit, mcopy[0].Fit
	var besti SpiderInfo = SpiderInfo{male, 0}
	for i := range mcopy {
		if worst.GetFit() > mcopy[i].Fit.GetFit() {
			worst = mcopy[i].Fit
		}
		if best.GetFit() < mcopy[i].Fit.GetFit() {
			best = mcopy[i].Fit
			besti = SpiderInfo{male, i}
		}
	}
	for i := range fcopy {
		if worst.GetFit() > fcopy[i].Fit.GetFit() {
			worst = fcopy[i].Fit
		}
		if best.GetFit() < fcopy[i].Fit.GetFit() {
			best = fcopy[i].Fit
			besti = SpiderInfo{female, i}
		}
	}
	for i := range mcopy {
		mcopy[i].Weight = (mcopy[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
	for i := range fcopy {
		fcopy[i].Weight = (fcopy[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
	med=mcopy.Median().Weight
	dominants:=mcopy.Filter(
		func(s Spider) bool {
			return s.Weight>med
		})
	var mbrood Spiders
	var fbrood Spiders
	for i := range dominants {
		fInRange := fcopy.Filter(
			func(s Spider) bool {
				return math.Abs(dominants[i].Passanger.Distance(s.Passanger))<=radius
			})
		if len(fInRange) < 1 {
			continue
		}
		f:=fInRange.SelectRandom()
		ch1, ch2, g1, g2:=dominants[i].Mate(f)
		
		if g1 == male {
			mbrood = append(mbrood, ch1)
		} else {
			fbrood = append(fbrood, ch1)
		}
		if g2 == male {
			mbrood = append(mbrood, ch2)
		} else {
			fbrood = append(fbrood, ch2)
		}
	}
	
	for i := range fbrood {
		res:=alg.fitF(fbrood[i].Passanger)
		fbrood[i].Fit=res
	}
	for i := range mbrood {
		res:=alg.fitF(mbrood[i].Passanger)
		mbrood[i].Fit=res
	}
	
	if len(fbrood) > 0 {
		fbrood = fbrood.Select(int(float64(len(alg.Females))*sChance))
		fn:=len(fcopy)-len(fbrood)
		fend:=fcopy.FitSort()[fn:]
		fend =append(fend, fbrood...).FitSort()
		fres :=fend[:len(fbrood)]
		fcopy =fcopy.FitSort()[:fn]
		fcopy = append(fcopy, fres...)
	}
	if len(mbrood) > 0 {
		mbrood = mbrood.Select(int(float64(len(alg.Males))*sChance))
		mn:=len(mcopy)-len(mbrood)
		mend := mcopy.FitSort()[mn:]
		mend = append(mend, mbrood...).FitSort()
		mres := mend[:len(mbrood)]
		mcopy = mcopy.FitSort()[:mn]
		mcopy = append(mcopy, mres...)
	}
	
	for i := range mcopy {
		if worst.GetFit() > mcopy[i].Fit.GetFit() {
			worst = mcopy[i].Fit
		}
		if best.GetFit() < mcopy[i].Fit.GetFit() {
			best = mcopy[i].Fit
			besti = SpiderInfo{male, i}
			if alg.BestFit.GetFit() <= best.GetFit() {
				alg.BestFit = best
				alg.BestPos.Copy(mcopy[i].Passanger)
			}
		}
	}
	for i := range fcopy {
		if worst.GetFit() > fcopy[i].Fit.GetFit() {
			worst = fcopy[i].Fit
		}
		if best.GetFit() < fcopy[i].Fit.GetFit() {
			best = fcopy[i].Fit
			besti = SpiderInfo{female, i}
			if alg.BestFit.GetFit() <= best.GetFit() {
				alg.BestFit = best
				alg.BestPos.Copy(fcopy[i].Passanger)
			}
		}
	}
	alg.worst=worst
	alg.best=best
	alg.besti=besti
	
	for i := range mcopy {
		mcopy[i].Weight = (mcopy[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
	for i := range fcopy {
		fcopy[i].Weight = (fcopy[i].Fit.GetFit()-worst.GetFit())/(best.GetFit()-worst.GetFit())
	}
	
	alg.Females = fcopy
	alg.Males = mcopy
}

func (alg SSO) New() EA {
	new := SSO{Spiders{},Spiders{}, nil, XInf, MInf, SpiderInfo{}, nil, MInf}
	return &new
}

func (alg SSO) Find(info SpiderInfo) Spider {
	if info.t == male {
		return alg.Males[info.i]
	} else if info.t == female {
		return alg.Females[info.i]
	} else {
		return Spider{}
	}
}

func (s Spider) FindBetterClosest(src Spiders) Spider {
	filtered:=src.Filter(
		func(s2 Spider) bool {
			return s2.Weight > s.Weight
		})
	res := filtered[0]
	d := s.Passanger.Distance(res.Passanger)
	for i := range filtered {
		d2:=s.Passanger.Distance(filtered[i].Passanger)
		if d > d2 {
			d=d2
			res=filtered[i]
		}
	}
	return res
}

func (s Spider) FindClosest(src Spiders) Spider {
	res := src[0]
	d := s.Passanger.Distance(res.Passanger)
	for i := range src {
		d2:=s.Passanger.Distance(src[i].Passanger)
		if d > d2 {
			d=d2
			res=src[i]
		}
	}
	return res
}

func (s Spider) VibWith(s2 Spider) float64 {
	d:=s.Passanger.Distance(s2.Passanger)
	return s2.Weight*math.Exp(-d*d)
}

func (s Spider) Mate(s2 Spider) (Spider, Spider, gender, gender) {
	var ch1, ch2 Spider
	ch1.Passanger,ch2.Passanger = s.Passanger.Crossover(s2.Passanger)
	ch1.Fit, ch2.Fit = MInf, MInf
	ch1.Weight, ch2.Weight = math.Inf(-1), math.Inf(-1)
	ch1.Passanger.Mutate(SSOmutProb, SSLowerB, SSUpperB)
	ch2.Passanger.Mutate(SSOmutProb, SSLowerB, SSUpperB)
	var g1,g2 gender
	if rand.Float64() < 0.5 {
		g1 = male
	} else {
		g1 = female
	}
	if rand.Float64() < 0.5 {
		g2 = male
	} else {
		g2 = female
	}
	return ch1, ch2, g1, g2
}

func (ss Spiders) Filter(f FilterF) Spiders {
	var new Spiders
	for i := range ss {
		if f(ss[i]) {
			new = append(new, ss[i])
		}
	}
	return new
}

func (ss Spiders) Select(n int) Spiders {
	var new Spiders
	var sum float64
	for _, ind := range ss {
		sum += ind.Fit.GetFit()
	}
	for i := 0; i<n-1; i++ {
		var k int
		pvar := rand.Float64()
		var psum float64
		for j, ind := range ss {
			psum += ind.Fit.GetFit()/sum
			if psum >= pvar {
				k = j
				break
			}
		}
		new = append(new, ss[k])
	}
	return new
}

func (ss Spiders) SelectBest(n int) Spiders {
	new:=ss.FitSort()
	return new[:len(new)/2]
}

func (ss Spiders) Median() Spider {
	res := ss.Sort()
	i := len(res)/2
	return res[i]
}

func (ss Spiders) Sort() Spiders {
	if len(ss) == 2 {
		if ss[0].Weight >= ss[1].Weight {
			new := make(Spiders, len(ss))
			copy(new, ss)
			return new
		} else {
			return Spiders{ss[1], ss[0]}
		}
	} else if len(ss) < 2 {
		new := make(Spiders, len(ss))
		copy(new, ss)
		return new
	}
	
	i := len(ss)/2
	var left, right Spiders
	for j := range ss {
		if i == j {continue}
		if ss[j].Weight >= ss[i].Weight {
			left = append(left, ss[j])
		} else {
			right = append(right, ss[j])
		}
	}
	left = left.Sort()
	right = right.Sort()
	return append(append(left, ss[i]), right...)
}

func (ss Spiders) FitSort() Spiders {
	if len(ss) == 2 {
		if ss[0].Fit.GetFit() >= ss[1].Fit.GetFit() {
			new := make(Spiders, len(ss))
			copy(new, ss)
			return new
		} else {
			return Spiders{ss[1], ss[0]}
		}
	} else if len(ss) < 2 {
		new := make(Spiders, len(ss))
		copy(new, ss)
		return new
	}
	
	i := len(ss)/2
	var left, right Spiders
	for j := range ss {
		if i == j {continue}
		if ss[j].Fit.GetFit() >= ss[i].Fit.GetFit() {
			left = append(left, ss[j])
		} else {
			right = append(right, ss[j])
		}
	}
	left = left.FitSort()
	right = right.FitSort()
	return append(append(left, ss[i]), right...)
}

func (ss Spiders) Mean() Individual {
	sum1 := ss[0].Passanger.Mult(ss[0].Weight)
	sum2 := ss[0].Weight
	for i := 1; i < len(ss); i++ {
		sum1 = sum1.Add(ss[i].Passanger.Mult(ss[i].Weight))
		sum2 += ss[i].Weight
	}
	return sum1.Div(sum2)
}

func (ss Spiders) SelectRandom() Spider {
	var sum float64
	for _, sp := range ss {
		sum += sp.Fit.GetFit()
	}
	var k int
	p := rand.Float64()
	var psum float64
	for i, sp := range ss {
		psum += sp.Fit.GetFit()/sum
		if psum > p {
			k = i
			break
		}
	}
	return ss[k]
}
