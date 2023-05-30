package EA

type (
	EA interface {
		StartUp(Individual, FitFunc, int)
		Iterate()
		New() EA
		GetBest() (ContextFit, Individual)
		GetCurrFits() []ContextFit
		GetType() string
		ResetTop()
	}
	Individual interface {
		Crossover(Individual) (Individual, Individual)
		Mutate(prob float64, low, up float64)
		Add(Individual) Individual
		GenAdd(float64) Individual
		Sub(Individual) Individual
		Mult(float64) Individual
		Div(float64) Individual
		PointwiseMult(Individual) Individual
		Random(float64, float64) Individual
		Get(int) float64
		Set(int, float64)
		Len() int
		Copy(Individual)
		Distance(Individual) float64
	}
	
	ContextFit interface {
		GetFit() float64
		Reset() ContextFit
		Copy() ContextFit
	}
	
	FitFunc func(Individual) ContextFit
)

const SSLowerB float64 = -200
const SSUpperB float64 = 200
var MInf ContextFit
var XInf ContextFit

func New[EAT EA] (alg EAT) EAT {
	return alg.New().(EAT)
}

func GenPopulation(ind Individual, size int) []Individual {
	pop := make([]Individual, size)
	for i := range pop {
		pop[i] = ind.Random(SSLowerB, SSUpperB)
	}
	return pop
}

