package main

import(
	"os"
	"strings"
	"strconv"
	"math"
	"fmt"
	"path/filepath"
	
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"AI/EA"
)

func main() {//generate graphs
	for _, ea := range []string{"GA", "PSO", "ABC", "SSO"} {
		for _, nn := range []string{"FNN", "RNN", "LSTM"} {
			var stratsStr map[string]string =map[string]string{"AVG":"","MIN":"","MAX":""}
			for _, typ := range []string{"AVG", "MIN", "MAX"} {
				f, _ := os.Open(filepath.Join(".","Stats","EA["+ea+"]NN["+nn+"]",typ+".stats.csv"))
				st,_:=f.Stat()
				data := make([]byte, st.Size())
				f.Read(data)
				str:=stratsStr[typ]
				str+=string(data)
				stratsStr[typ]=str
				f.Close()
			}
			var stratsData map[string]map[string][]float64=map[string]map[string][]float64{"AVG":nil,"MIN":nil,"MAX":nil}
			for typ := range stratsStr {
				str:=stratsStr[typ]
				strarr:=strings.Split(str, ";")
				strarr=strarr[:len(strarr)-1]
				var strmatr [][]string
				for i:= range strarr {
					strmatr=append(strmatr,strings.Split(strarr[i], ","))
				}
				var matr map[string][]float64 = map[string][]float64{"Fit":nil,"Apples Eaten":nil,"Kills":nil,"Time Survived":nil,"Penalty":nil/*,"PenPSec":nil*/}
				converter := []string{"Fit","ApplesEaten","Kills","TimeSurvived","Penalty"}
				_=converter
				for i:= range strmatr {
					for j:= range strmatr[i] {
						stat := converter[j]
						f,_:=strconv.ParseFloat(strmatr[i][j],64)
						if stat == "TimeSurvived" {
							matr[stat] = append(matr[stat], f/1e9)
						} else {
							matr[stat] = append(matr[stat], f)
						}
					}
				}
				stratsData[typ]=matr
			}
			
			for _, stat := range []string{"Fit","ApplesEaten","Kills","TimeSurvived","Penalty"/*,"PenPSec"*/} {
				p := plot.New()
				p.Title.Text = "EA:"+ea+";NN:"+nn+";Stat:"+stat
				p.X.Label.Text = "Iteration"
				p.Y.Label.Text = ToText(stat)
				plotutil.AddLinePoints(p,"Maximum", ConvToPlot(stratsData["MAX"][stat]),"Avarage", ConvToPlot(stratsData["AVG"][stat]),"Minimum", ConvToPlot(stratsData["MIN"][stat]))
				p.Save(8*vg.Inch, 8*vg.Inch , "EA["+ea+"]NN["+nn+"]."+stat+".jpg")
			}
		}
	}
}

type CF struct {
	a float64
}

func (cf CF) GetFit() float64 {
	return cf.a
}

func (cf CF) Reset() EA.ContextFit {
	return cf
}

func (cf CF) Copy() EA.ContextFit {
	return cf
}

func main2() {//calculate maximums and aggregate
	var Data map[string]map[string]float64 = make(map[string]map[string]float64,0)
	for _, ea := range []string{"GA", "PSO", "ABC", "SSO"} {
		for _, nn := range []string{"FNN", "RNN", "LSTM"} {
			var max float64 = math.Inf(-1)
			f, _ := os.Open(filepath.Join(".","Stats","EA["+ea+"]NN["+nn+"]","MAX.stats.csv"))
			st,_:=f.Stat()
			data := make([]byte, st.Size())
			f.Read(data)
			str:=string(data)
			f.Close()
			strarr:=strings.Split(str, ";")
			strarr=strarr[:len(strarr)-1]
			for i:= range strarr {
				a1, _:=strconv.ParseFloat(strings.Split(strarr[i], ",")[4],64)
				a2,_:=strconv.ParseFloat(strings.Split(strarr[i], ",")[3],64)
				a:=a1/(a2/1e9)//Penalty per second
				//a, _:=strconv.ParseFloat(strings.Split(strarr[i], ",")[4],64)//Fit, Apples Eaten, Kills, TimeSurv (ns, need to /1e9), Penalty
				if max < a {
					max=a
				}
			}
			if v:=Data[ea]; v==nil {
				Data[ea]=make(map[string]float64,0)
			}
			Data[ea][nn]=max
			fmt.Printf("%s %s: %.2f\n", ea, nn, max)
		}
	}
	for _, ea := range []string{"GA", "PSO", "ABC", "SSO"} {
		sum := 0.0
		l := len(Data[ea])
		for nn := range Data[ea] {
			sum+=Data[ea][nn]
		}
		fmt.Printf("%s(%d): %.2f\n", ea, l, sum/float64(l))
	}
	
	for _, nn := range []string{"FNN", "RNN", "LSTM"} {
		sum := 0.0
		l := len(Data)
		for ea := range Data {
			sum+=Data[ea][nn]
		}
		fmt.Printf("%s(%d): %.2f\n", nn, l, sum/float64(l))
	}
}

func ConvToPlot(arr []float64) plotter.XYs {
	pts := make(plotter.XYs, len(arr))
	for i := range arr {
		pts[i].X = float64(i)
		pts[i].Y = arr[i]
	}
	return pts
}

func ToText(stat string) string {
	switch stat {
		case "Fit": return "fit"
		case "ApplesEaten": return "apples"
		case "Kills": return "kills"
		case "TimeSurvived": return "sec"
		case "Penalty": return "incorrect moves"
		case "PenPSec": return "penalty per second"
		default: return "y"
	}
}
