package main

import (
	"fmt"
	"log"
	"AI/NN"
	"AI/EA"
	"math/rand"
	"math"
	"time"
	"encoding/json"
	"os"
	"sync"
	"path/filepath"
)

type(
	DB struct {
		Snakes map[string]SnakeData
		Apples []Coord
		MapSize Coord
		mutex chan bool
	}
	SnakeData struct {
		Head Coord
		Body []Coord
		OldTail Coord
		Dir Coord
		ApplesEaten int
		Kills int
		Started time.Time
		Alive bool
		Penalty int
	}
	Coord struct {
		x, y int
	}
	CommandLine struct {
		cmd Command
		login string
		dir Coord
	}
	Command int
	DataSet []Data
	Data []Coord
)

var GlobalDB DB
const (
	register Command = 5<<iota
	spawn
	do
	acknowledge
	bye
)
const appleChance float64 = 0.1

const PreTrainingTime float64 = 12
const TrainingTime float64 = 12

var epsilon float64
var epsiLocker sync.RWMutex

func main() {
	rand.Seed(int64(time.Now().Nanosecond()))
	EA.MInf = ContextFit{math.Inf(-1),0,0,0,0}
	EA.XInf = ContextFit{math.Inf(1),0,0,0,0}
	CommChansS := make(map[string](chan bool))
	CommChansR := make(map[string](chan CommandLine))
	NNs := []NN.NN{NN.FNNtempl, NN.RNNtempl, NN.LSTMtempl}
	EAs := []EA.EA{EA.GAtempl, EA.PSOtempl, EA.ABCtempl, EA.SSOtempl}
	ChangeModeComm := make([]chan<- bool,0)
	Acknowledges := make([]<-chan string,0)
	for _, nn := range NNs {
		for _, alg := range EAs {
			var localDB DB
			localDB.Snakes = make(map[string]SnakeData)
			localDB.Apples = make([]Coord, 0)
			localDB.MapSize = Coord{100, 100}
			localDB.mutex = make(chan bool, 1)
			localDB.mutex <- true
			var gameChanReceive chan bool = make(chan bool,1)
			var gameChanSend chan CommandLine = make(chan CommandLine,1)
			var optChanReceive chan NN.Individual = make(chan NN.Individual, 1)
			var optChanSend chan NN.Vector = make(chan NN.Vector, 1)
			go localHandler(&localDB, gameChanReceive, gameChanSend)
			fitF := fitGen(optChanReceive, optChanSend)
			key:=genUniqueKey(alg,nn)
			var gameChanReceive2 chan bool = make(chan bool,1)
			var gameChanSend2 chan CommandLine = make(chan CommandLine,1)
			CommChansS[key] = gameChanReceive2
			CommChansR[key] = gameChanSend2
			var gcr *<-chan bool = new(<-chan bool)
			var gcs *chan<- CommandLine = new(chan<- CommandLine)
			*gcr = gameChanReceive
			*gcs = gameChanSend
			cmc := make(chan bool, 1)
			ChangeModeComm = append(ChangeModeComm, cmc)
			CommWithA := make(chan bool)
			var params []int
			switch nn.(type) {
				case *NN.LSTM:
					params=[]int{40,4}
				default:
					params=[]int{40,20,4}
			}
			LastCh := make(chan NN.Individual, 1)
			AckCh := make(chan string, 1)
			Acknowledges = append(Acknowledges, AckCh)
			go optimizer(nn, alg, fitF, cmc, CommWithA, gcr, gcs, gameChanReceive2, gameChanSend2, LastCh, AckCh, params...)
			go gameAgent(&localDB, key, nn, gcr, gcs, optChanReceive, optChanSend, CommWithA, LastCh, params...)
		}
	}
	log.Println("Timer start")
	go epsilonController(0.5)
	time.Sleep(time.Duration(PreTrainingTime)*time.Hour)
	log.Println("Switch start")
	GlobalDB.Snakes = make(map[string]SnakeData, 0)
	GlobalDB.Apples = make([]Coord, 0)
	GlobalDB.MapSize = Coord{100, 100}
	GlobalDB.mutex = make(chan bool, 1)
	GlobalDB.mutex <- true
	go gameHandler(&CommChansS, &CommChansR)
	for i := range ChangeModeComm {
		j:=i
		go func() {ChangeModeComm[j]<-true}()
	}
	time.Sleep(time.Duration(TrainingTime)*time.Hour)
	log.Println("Start ending sequence")
	for i := range ChangeModeComm {
		ChangeModeComm[i]<-true
	}
	
	for _, ch := range Acknowledges {
		log.Println(<-ch,"is done")
	}
	log.Println("Ending....")
	time.Sleep(2*time.Minute)
	log.Println("End")
}

func epsilonController(startEpsilon float64) {
	TotalTime := PreTrainingTime + TrainingTime
	for i:=1.0; i <= TotalTime; i++{
		time.Sleep(1*time.Hour)
		epsiLocker.Lock()
			epsilon = startEpsilon*math.Pow(0.1,i/TotalTime)
		epsiLocker.Unlock()
	}
}

func (c1 Coord) Distance(c2 Coord) float64 {
	dx, dy := c2.x-c1.x, c2.y-c1.y
	if dx == 0 {
		return math.Abs(float64(dy))
	} else if dy == 0 {
		return math.Abs(float64(dx))
	} else {
		return math.Sqrt(float64(dx*dx+dy*dy))
	}
}

func (c Coord) IsIn(min Coord, max Coord) bool {
	return c.x>=min.x && c.x<=max.x && c.y>=min.y && c.y<=max.y
}

func (d Data) Distance(c Coord, dir int, hugeVal float64) float64 {
	var filter func(Coord) bool
	switch dir {
		case 0:filter = func(c2 Coord) bool {return c2.x==c.x && c2.y<c.y}
		case 1:filter = func(c2 Coord) bool {dx:=c2.x-c.x;dy:=c2.y-c.y;return dx>0 && dy<0 && dx==-dy}
		case 2:filter = func(c2 Coord) bool {return c2.x>c.x && c2.y==c.y}
		case 3:filter = func(c2 Coord) bool {dx:=c2.x-c.x;dy:=c2.y-c.y;return dx>0 && dy>0 && dx==dy}
		case 4:filter = func(c2 Coord) bool {return c2.x==c.x && c2.y>c.y}
		case 5:filter = func(c2 Coord) bool {dx:=c2.x-c.x;dy:=c2.y-c.y;return dx<0 && dy>0 && -dx==dy}
		case 6:filter = func(c2 Coord) bool {return c2.x<c.x && c2.y==c.y}
		case 7:filter = func(c2 Coord) bool {dx:=c2.x-c.x;dy:=c2.y-c.y;return dx<0 && dy<0 && dx==dy}
		default:filter = func(c2 Coord) bool {return false}
	}
	var filteredDist []float64
	for i:= range d {
		if filter(d[i]) {
			filteredDist = append(filteredDist, c.Distance(d[i]))
		}
	}
	if len(filteredDist)<1 {
		return hugeVal
	}else {
		min := filteredDist[0]
		for i:=1; i<len(filteredDist); i++ {
			if min>filteredDist[i] {
				min=filteredDist[i]
			}
		}
		return min
	}
}

func (db *DB) GetDataWithMutex(login string) (DataSet, SnakeData) {
	<-db.mutex
	snake := db.Snakes[login].Copy()
	data := db.CopyAsDataSet(login)
	db.mutex<-true
	data[0] = append(data[0],snake.Body...)
	return data, snake
}

func (db DB) CopyAsDataSet(login string) DataSet {
	data:=make(DataSet,4)
	data[1] = append(data[1], db.Apples...)
	for k:= range db.Snakes{
		if k != login {
			data[2] = append(data[2], db.Snakes[k].Head)
			data[3] = append(data[3], db.Snakes[k].Body...)
		}
	}
	return data
}

func (db DB) Copy() DB {
	var newDB DB
	newDB.Apples = []Coord{}
	newDB.Apples = append(newDB.Apples, db.Apples...)
	newDB.Snakes = make(map[string]SnakeData)
	for k:= range db.Snakes{
		s := db.Snakes[k]
		s.Body = append([]Coord{}, db.Snakes[k].Body...)
		newDB.Snakes[k] = s
	}
	newDB.MapSize = db.MapSize
	newDB.mutex = make(chan bool, 1)
	return newDB
}

func Min(a, b int) int {
	if a<b {
		return a
	} else {
		return b
	}
}

func Max(a, b int) int {
	if a>b {
		return a
	} else {
		return b
	}
}

func (db DB) CheckObjExistance(min, max Coord) bool {
	for _, a := range db.Apples {
		if a.IsIn(min,max) {
			return true
		}
	}
	for _, s := range db.Snakes {
		if s.Head.IsIn(min,max) {
			return true
		}
		for _, b := range s.Body {
			if b.IsIn(min,max) {
				return true
			}
		}
	}
	return false
}

func (s SnakeData) Copy() SnakeData {
	var sd SnakeData
	sd.Head = s.Head
	sd.Body = append(sd.Body, s.Body...)
	sd.Dir = s.Dir
	sd.ApplesEaten = s.ApplesEaten
	sd.Kills = s.Kills
	sd.Started = s.Started
	sd.Alive = s.Alive
	sd.Penalty = s.Penalty
	return sd
}

func gameHandler(tickSend *map[string](chan bool), cmdRec *map[string](chan CommandLine)) {
	log.Println("Started main game handler")
	for {
		t:= time.Now()
		forProcessing := make(map[string]CommandLine)
		for k:= range *cmdRec {
			if len((*cmdRec)[k])!=0 {
				forProcessing[k]=<-(*cmdRec)[k]
			}
		}
		dbcopy:=GlobalDB.Copy()
		for k:= range forProcessing {
			switch forProcessing[k].cmd {
				case do:
					s:=dbcopy.Snakes[k]
					if s.Dir.x!=forProcessing[k].dir.x && s.Dir.y!=forProcessing[k].dir.y {//perpendicular dir change
						s.Dir=forProcessing[k].dir
						dbcopy.Snakes[k] = s
					} else if s.Dir.x==forProcessing[k].dir.x && s.Dir.y==forProcessing[k].dir.y {// no dir change
					} else {//go in opposite dir
						s.Penalty++
						dbcopy.Snakes[k] = s
					}
				case register:
					dbcopy.Snakes[k] = SnakeData{Coord{}, nil, Coord{}, Coord{}, 0,0, time.Now(), false, 0}
				case spawn:
HeadRep:
					head:=Coord{rand.Intn(dbcopy.MapSize.x-2)+1, rand.Intn(dbcopy.MapSize.y-2)+1}
					dir:=Coord{}
Repeat:
					p:=rand.Float64()
					switch {
						case p<0.25 && head.y+2<dbcopy.MapSize.y: dir=Coord{0,-1}
						case p<0.50 && head.x-2>=0: dir=Coord{1,0}
						case p<0.75 && head.y-2>=0: dir=Coord{0,1}
						case p<1.00 && head.x+2<dbcopy.MapSize.x: dir=Coord{-1,0}
						default: goto Repeat
					}
					body:=[]Coord{Coord{head.x-dir.x,head.y-dir.y}, Coord{head.x-2*dir.x,head.y-2*dir.y}}
					min:=Coord{Min(head.x, body[1].x)-1, Min(head.y, body[1].y)-1}
					max:=Coord{Max(head.x, body[1].x)+1, Max(head.y, body[1].y)+1}
					if dbcopy.CheckObjExistance(min, max) {goto HeadRep}
					//move back, so during snake move step it will be iterated back
					head.x -= dir.x
					head.y -= dir.y
					body[0].x -= dir.x
					body[0].y -= dir.y
					body[1].x -= dir.x
					body[1].y -= dir.y
					dbcopy.Snakes[k] = SnakeData{head, body, Coord{}, dir, 0,0, time.Now(), true, 0}
			}
		}
		//move snakes
		for k, s := range dbcopy.Snakes {
			if s.Alive {
				old:=s.Head
				s.Head=Coord{old.x+s.Dir.x, old.y+s.Dir.y}
				s.OldTail = s.Body[len(s.Body)-1]
				s.Body=append([]Coord{old},s.Body[:len(s.Body)-1]...)
				dbcopy.Snakes[k]=s
			}
		}
		//check and eat || death
SnakesLoop:
		for k, s := range dbcopy.Snakes {
			if s.Alive {
				if s.Head.x<0 || s.Head.x>=dbcopy.MapSize.x || s.Head.y<0 || s.Head.y>=dbcopy.MapSize.y {
					s.Alive=false
					dbcopy.Snakes[k]=s
					continue
				}
				for _, b := range s.Body {
					if s.Head==b {
						s.Alive=false
						dbcopy.Snakes[k]=s
						continue SnakesLoop
					}
				}
				for i, a := range dbcopy.Apples {
					if a==s.Head {
						la:=len(dbcopy.Apples)
						if i < la-1 {
							dbcopy.Apples = append(dbcopy.Apples[:i], dbcopy.Apples[i+1:]...)
						} else {
							dbcopy.Apples = dbcopy.Apples[:i]
						}
						s.ApplesEaten++
						s.Body = append(s.Body, s.OldTail)
						dbcopy.Snakes[k]=s
						break
					}
				}
				for k2, s2 := range dbcopy.Snakes {
					if k!=k2 && s2.Alive {
						if s.Head==s2.Head {
							s.Alive=false
							s2.Alive=false
							dbcopy.Snakes[k]=s
							dbcopy.Snakes[k2]=s2
							break
						}
						for _, b := range s2.Body {
							if s.Head==b {
								s.Alive=false
								s2.Kills++
								dbcopy.Snakes[k]=s
								dbcopy.Snakes[k2]=s2
								break
							}
						}
					}
				}
			}
		}
		//spawn apple
		p:=rand.Float64()
		if len(dbcopy.Apples) < 10 && p<appleChance {
AppleRep:
			newA :=Coord{rand.Intn(dbcopy.MapSize.x), rand.Intn(dbcopy.MapSize.y)}
			for _, a := range dbcopy.Apples {
				if a==newA {
					goto AppleRep
				}
			}
			for _, s := range dbcopy.Snakes {
				if !s.Alive {
					continue
				}
				if s.Head == newA {
					goto AppleRep
				}
				for _, b := range s.Body {
					if b==newA {
						goto AppleRep
					}
				}
			}
			dbcopy.Apples=append(dbcopy.Apples,newA)
		}
		//modify actual db
		<-GlobalDB.mutex
		GlobalDB.Apples=append([]Coord{},dbcopy.Apples...)
		for k, s := range dbcopy.Snakes {
			GlobalDB.Snakes[k]=s
		}
		GlobalDB.mutex<-true
		//send ticks
		for k := range forProcessing {
			(*tickSend)[k]<-true
		}
		
		//rest of the tick
		l:=time.Since(t).Nanoseconds()
		if l < 100*int64(time.Millisecond) {
			time.Sleep(time.Duration(100*int64(time.Millisecond)-time.Since(t).Nanoseconds()))
		}
	}
	
}

func localHandler(ldb *DB, lhs chan<- bool, lhr <-chan CommandLine) {
	for {
		t:=time.Now()
		dbcopy:=ldb.Copy()
		rec := false
		select {
			case cmd:=<-lhr:
				rec = true
				switch cmd.cmd {
					case bye:
						log.Println(cmd.login, "stopped local handler")
						time.Sleep(time.Minute)
						return
					case do:
						s:=dbcopy.Snakes[cmd.login]
						if s.Dir.x!=cmd.dir.x && s.Dir.y!=cmd.dir.y {//perpendicular dir change
							s.Dir=cmd.dir
							dbcopy.Snakes[cmd.login] = s
						} else if s.Dir.x==cmd.dir.x && s.Dir.y==cmd.dir.y {// no dir change
						} else {//go in opposite dir
							s.Penalty++
							dbcopy.Snakes[cmd.login] = s
						}
					case register:
						dbcopy.Snakes[cmd.login] = SnakeData{Coord{}, []Coord{}, Coord{}, Coord{}, 0,0, time.Now(), false, 0}
					case spawn:
						head:=Coord{rand.Intn(dbcopy.MapSize.x-2)+1, rand.Intn(dbcopy.MapSize.y-2)+1}
						dir:=Coord{}
Repeat:
						p:=rand.Float64()
						switch {
							case p<0.25 && head.y+2<dbcopy.MapSize.y: dir=Coord{0,-1}
							case p<0.50 && head.x-2>=0: dir=Coord{1,0}
							case p<0.75 && head.y-2>=0: dir=Coord{0,1}
							case p<1.00 && head.x+2<dbcopy.MapSize.x: dir=Coord{-1,0}
							default: goto Repeat
						}
						body:=[]Coord{Coord{head.x-dir.x,head.y-dir.y}, Coord{head.x-2*dir.x,head.y-2*dir.y}}
						dbcopy.Snakes[cmd.login] = SnakeData{head, body, Coord{}, dir, 0,0, time.Now(), true, 0}
				}
			default:
		}
		//move snakes
		for k, s := range dbcopy.Snakes {
			if s.Alive {
				old:=s.Head
				s.Head=Coord{old.x+s.Dir.x, old.y+s.Dir.y}
				s.OldTail = s.Body[len(s.Body)-1]
				s.Body=append([]Coord{old},s.Body[:len(s.Body)-1]...)
				dbcopy.Snakes[k]=s
			}
		}
		//check and eat || death
SnakesLoop:
		for k, s := range dbcopy.Snakes {
			if s.Alive {
				if s.Head.x<0 || s.Head.x>=dbcopy.MapSize.x || s.Head.y<0 || s.Head.y>=dbcopy.MapSize.y {
					s.Alive=false
					dbcopy.Snakes[k]=s
					break
				}
				for _, b := range s.Body {
					if s.Head==b {
						s.Alive=false
						dbcopy.Snakes[k]=s
						break SnakesLoop
					}
				}
				for i, a := range dbcopy.Apples {
					if a==s.Head {
						la:=len(dbcopy.Apples)
						if i < la-1 {
							dbcopy.Apples = append(dbcopy.Apples[:i], dbcopy.Apples[i+1:]...)
						} else {
							dbcopy.Apples = dbcopy.Apples[:i]
						}
						s.ApplesEaten++
						s.Body = append(s.Body, s.OldTail)
						dbcopy.Snakes[k]=s
						break
					}
				}
			}
		}
		//spawn apple
		p:=rand.Float64()
		if len(dbcopy.Apples) < 5 && p<appleChance {
AppleRep:
			newA :=Coord{rand.Intn(dbcopy.MapSize.x), rand.Intn(dbcopy.MapSize.y)}
			for _, a := range dbcopy.Apples {
				if a==newA {
					goto AppleRep
				}
			}
			for _, s := range dbcopy.Snakes {
				if !s.Alive {
					break
				}
				if s.Head == newA {
					goto AppleRep
				}
				for _, b := range s.Body {
					if b==newA {
						goto AppleRep
					}
				}
			}
			dbcopy.Apples=append(dbcopy.Apples,newA)
		}
		//modify actual db
		<-ldb.mutex
		ldb.Apples=append([]Coord{},dbcopy.Apples...)
		for k, s := range dbcopy.Snakes {
			ldb.Snakes[k]=s
		}
		ldb.mutex<-true
		//send ticks
		if rec {
			lhs<-true
		}
		l:=time.Since(t).Nanoseconds()
		if l < 100*int64(time.Millisecond) {
			time.Sleep(time.Duration(100*int64(time.Millisecond)-time.Since(t).Nanoseconds()))
		}
	}
}

func gameAgent(db *DB, key string, nnt NN.NN, gameChanReceive *(<-chan bool), gameChanSend *(chan<- CommandLine), optChanReceive <-chan NN.Individual, optChanSend chan<- NN.Vector, switchChan <-chan bool, FinalChan <-chan NN.Individual, params ...int) {
	log.Println(key, "started agent")
	var inGame bool
	nn,_:=NN.New(nnt,params...)
	*gameChanSend<-CommandLine{register, key, Coord{}}
	HugeVal:=float64(db.MapSize.x*db.MapSize.y+100000)
	old:=*gameChanSend
	var ind NN.Individual
Loop:
	for {
		select {
			case <-*gameChanReceive:
			if !inGame {
Rep:
				select {
					case ind=<-FinalChan:
						log.Println(key, "agent switched into non-trainig mode")
						*gameChanSend<-CommandLine{acknowledge, key, Coord{}}
						break Loop
					case <-switchChan:
						<-switchChan
						db=&GlobalDB
						*gameChanSend<-CommandLine{register, key, Coord{}}
						old<-CommandLine{bye, key, Coord{}}
					case ind:=<-optChanReceive:
						nn.GenFromIndividual(ind)
						inGame=true
						*gameChanSend<-CommandLine{spawn, key, Coord{}}
					default: goto Rep
				}
			}else {
				DataArr, self :=db.GetDataWithMutex(key)
				if self.Alive {
					var v NN.Vector
					for i:=0; i<8; i++ {//dirs
						for j:=0; j < 5; j++ {//types
							if j!=0 {
								dist := DataArr[j-1].Distance(self.Head, i, HugeVal)
								v=append(v, dist)
							} else {
								switch i {
									case 0:v=append(v, float64(self.Head.y))
									case 1:v=append(v, self.Head.Distance(Coord{db.MapSize.x,0}))
									case 2:v=append(v, float64(db.MapSize.x-self.Head.x))
									case 3:v=append(v, self.Head.Distance(Coord{db.MapSize.x,db.MapSize.y}))
									case 4:v=append(v, float64(db.MapSize.y-self.Head.y))
									case 5:v=append(v, self.Head.Distance(Coord{0,db.MapSize.y}))
									case 6:v=append(v, float64(self.Head.x))
									case 7:v=append(v, self.Head.Distance(Coord{0,0}))
									default:v=append(v, 0)
								}
							}
						}
					}
					resV,_:=nn.Calculate(v)
					i:=SelectIndex(resV)
					var dir Coord
					switch i {
						case 0: dir = Coord{0,-1}//up
						case 1: dir = Coord{1,0}//right
						case 2: dir = Coord{0,1}//down
						case 3: dir = Coord{-1,0}//left
						default: dir = self.Dir//error
					}
					*gameChanSend<-CommandLine{do, key, dir}
				}else {
					inGame=false
					dur:=time.Since(self.Started)
					optChanSend<-NN.Vector{float64(self.ApplesEaten), float64(self.Kills), float64(dur), float64(self.Penalty)}//dur in nanosec
					*gameChanSend<-CommandLine{acknowledge, key, Coord{}}
				}
			}
			default:
		}
	}
	for {
		select {
			case <-*gameChanReceive:
				if !inGame {
						nn.GenFromIndividual(ind)
						inGame=true
						*gameChanSend<-CommandLine{spawn, key, Coord{}}
				}else {
					DataArr, self :=db.GetDataWithMutex(key)
					if self.Alive {
						var v NN.Vector
						for i:=0; i<8; i++ {//dirs
							for j:=0; j < 5; j++ {//types
								if j!=0 {
									dist := DataArr[j-1].Distance(self.Head, i, HugeVal)
									v=append(v, dist)
								} else {
									switch i {
										case 0:v=append(v, float64(self.Head.y))
										case 1:v=append(v, self.Head.Distance(Coord{db.MapSize.x,0}))
										case 2:v=append(v, float64(db.MapSize.x-self.Head.x))
										case 3:v=append(v, self.Head.Distance(Coord{db.MapSize.x,db.MapSize.y}))
										case 4:v=append(v, float64(db.MapSize.y-self.Head.y))
										case 5:v=append(v, self.Head.Distance(Coord{0,db.MapSize.y}))
										case 6:v=append(v, float64(self.Head.x))
										case 7:v=append(v, self.Head.Distance(Coord{0,0}))
										default:v=append(v, 0)
									}
								}
							}
						}
						resV,_:=nn.Calculate(v)
						i:=SelectIndex(resV)
						var dir Coord
						switch i {
							case 0: dir = Coord{0,-1}//up
							case 1: dir = Coord{1,0}//right
							case 2: dir = Coord{0,1}//down
							case 3: dir = Coord{-1,0}//left
							default: dir = self.Dir//error
						}
						*gameChanSend<-CommandLine{do, key, dir}
					}else {
						inGame=false
						*gameChanSend<-CommandLine{acknowledge, key, Coord{}}
					}
				}
			default:
		}
	}
}

func SelectIndex(vals NN.Vector) int {
	epsiLocker.RLock()
	eps:=epsilon
	epsiLocker.RUnlock()
	p:=rand.Float64()
	if p<=1-eps {
		max := vals[0]
		best:=0
		for i:=1; i<len(vals); i++ {
			if max<vals[i] {
				max=vals[i]
				best=i
			}
		}
		return best
	} else {
		return rand.Intn(len(vals))
	}	
}

func optimizer(nnt NN.NN, algt EA.EA, ff EA.FitFunc, timeToChangeState <-chan bool, informGameAgent chan<- bool, pointerA *<-chan bool, pointerB *chan<- CommandLine, valA chan bool, valB chan<- CommandLine, LastCh chan<- NN.Individual, AckCh chan<- string,params ...int) {
	key:=genUniqueKey(algt, nnt)
	log.Println(key, "started optimizer")
	alg:=EA.New(algt)
	ind:=nnt.New(params...).MakeIndividual()
	StColCh := make(chan []EA.ContextFit, 1)
	CollectorDoneCh := make(chan bool, 1)
	hasBeenSwitched := false
	go statcollector(key, StColCh, CollectorDoneCh)
	alg.StartUp(&ind, ff, 20)
	log.Println(key, "optimizer initialized")
	//StColCh<-alg.GetCurrFits()
Loop:
	for {
		alg.Iterate()
		if hasBeenSwitched {
			StColCh<-alg.GetCurrFits()
		}
		select {
			case <-timeToChangeState:
				if !hasBeenSwitched {
					informGameAgent<-true
					*pointerA=valA
					*pointerB=valB
					informGameAgent<-true
					hasBeenSwitched = true
					alg.ResetTop()
					log.Println(key, "agent switched into multiplayer mode")
				} else {
					log.Println(key, "started ending sequence")
					for len(StColCh) !=0 {}
					StColCh<-[]EA.ContextFit{}
					BstFit, BstInd := alg.GetBest()
					LastCh<- *(BstInd.(*NN.Individual))
					fin, err:=os.Create(filepath.Join("Trained",key+".final.json"))
					if err!=nil {
						log.Println(err)
					}
					var VInd BestIndividualJSON
					VInd.Ind.Copy(BstInd)
					VInd.Layers=params
					VInd.Type=nnt.GetType()
					VInd.Fit=BstFit.(ContextFit).ReplaceInfs()
					res, err:=json.Marshal(VInd)
					if err!=nil {
						log.Println(err)
					}
					fmt.Fprint(fin, string(res))
					fin.Close()
					<-CollectorDoneCh
					AckCh<-key
					log.Println(key, "stoped optimizer")
					break Loop
				}
			default:
		}
	}
}

type BestIndividualJSON struct {
	Type string
	Layers []int
	Ind NN.Individual
	Fit ContextFit
}

type StatsVar struct{
	Max ContextFit
	Avg ContextFit
	Min ContextFit
}

func statcollector(key string, ch chan []EA.ContextFit, Done chan bool) {
	log.Println(key, "started stats collector")
	var statsArr []StatsVar
	for {
		data:=<-ch
		if len(data) == 0 {
			break
		}
		var convData []ContextFit
		for i := range data {
			convData=append(convData, data[i].(ContextFit))
		}
		var max, sum, min ContextFit
		max, min = convData[0], convData[0]
		for _, v := range convData {
			sum=sum.Add(v)
			max = max.EachMax(v)
			min = min.EachMin(v)
		}
		avg := sum.Div(float64(len(data)))
		statsArr = append(statsArr,StatsVar{max,avg,min})
	}
	os.Mkdir(key, 0750)
	minfin, _ := os.Create(filepath.Join(key,"MIN.stats.csv"))
	maxfin, _ := os.Create(filepath.Join(key,"MAX.stats.csv"))
	avgfin, _ := os.Create(filepath.Join(key,"AVG.stats.csv"))
	defer minfin.Close()
	defer maxfin.Close()
	defer avgfin.Close()
	for _, v := range statsArr {
		fmt.Fprintf(minfin,"%.10f,%.10f,%.10f,%.10f,%.10f;",v.Min.Fit, v.Min.ApplesEaten, v.Min.Kills, v.Min.TimeSurv, v.Min.Penalty)
		fmt.Fprintf(maxfin,"%.10f,%.10f,%.10f,%.10f,%.10f;",v.Max.Fit, v.Max.ApplesEaten, v.Max.Kills, v.Max.TimeSurv, v.Max.Penalty)
		fmt.Fprintf(avgfin,"%.10f,%.10f,%.10f,%.10f,%.10f;",v.Avg.Fit, v.Avg.ApplesEaten, v.Avg.Kills, v.Avg.TimeSurv, v.Avg.Penalty)
	}
	Done<-true
	log.Println(key, "stopped stats collector")
}

type ContextFit struct {
	Fit float64
	ApplesEaten float64
	Kills float64
	TimeSurv float64//dur in nanosec
	Penalty float64
}

func (CF ContextFit) Reset() EA.ContextFit {
	return ContextFit{math.Inf(-1),0,0,0,0}
}

func (CF ContextFit) Copy() EA.ContextFit {
	var newCF ContextFit
	newCF.Fit=CF.Fit
	newCF.ApplesEaten=CF.ApplesEaten
	newCF.Kills=CF.Kills
	newCF.TimeSurv=CF.TimeSurv
	newCF.Penalty=CF.Penalty
	return newCF
}

func (CF ContextFit) ReplaceInfs() ContextFit {
	if math.IsInf(CF.Fit,0) {
		CF.Fit = math.Copysign(100000, CF.Fit)
	}
	if math.IsInf(CF.Fit,0) {
		CF.ApplesEaten = math.Copysign(100000, CF.ApplesEaten)
	}
	if math.IsInf(CF.Fit,0) {
		CF.Kills = math.Copysign(100000, CF.Kills)
	}
	if math.IsInf(CF.Fit,0) {
		CF.TimeSurv = math.Copysign(100000, CF.TimeSurv)
	}
	if math.IsInf(CF.Fit,0) {
		CF.Penalty = math.Copysign(100000, CF.Penalty)
	}
	return CF
}

func (CF ContextFit) GetFit() float64 {
	return CF.Fit
}

func (CF ContextFit) Add(v ContextFit) ContextFit {
	return ContextFit{CF.Fit+v.Fit, CF.ApplesEaten+v.ApplesEaten, CF.Kills+v.Kills,CF.TimeSurv+v.TimeSurv, CF.Penalty+v.Penalty}
}

func (CF ContextFit) Div(x float64) ContextFit {
	return ContextFit{CF.Fit/x, CF.ApplesEaten/x, CF.Kills/x,CF.TimeSurv/x, CF.Penalty/x}
}

func (a ContextFit) EachMax(b ContextFit) ContextFit {
	var res ContextFit
	if a.Fit > b.Fit {
		res.Fit=a.Fit
	} else {
		res.Fit=b.Fit
	}
	if a.ApplesEaten > b.ApplesEaten {
		res.ApplesEaten=a.ApplesEaten
	} else {
		res.ApplesEaten=b.ApplesEaten
	}
	if a.Kills > b.Kills {
		res.Kills=a.Kills
	} else {
		res.Kills=b.Kills
	}
	if a.TimeSurv > b.TimeSurv {
		res.TimeSurv=a.TimeSurv
	} else {
		res.TimeSurv=b.TimeSurv
	}
	if a.Penalty > b.Penalty {
		res.Penalty=a.Penalty
	} else {
		res.Penalty=b.Penalty
	}
	return res
}

func (a ContextFit) EachMin(b ContextFit) ContextFit {
	var res ContextFit
	if a.Fit < b.Fit {
		res.Fit=a.Fit
	} else {
		res.Fit=b.Fit
	}
	if a.ApplesEaten < b.ApplesEaten {
		res.ApplesEaten=a.ApplesEaten
	} else {
		res.ApplesEaten=b.ApplesEaten
	}
	if a.Kills < b.Kills {
		res.Kills=a.Kills
	} else {
		res.Kills=b.Kills
	}
	if a.TimeSurv < b.TimeSurv {
		res.TimeSurv=a.TimeSurv
	} else {
		res.TimeSurv=b.TimeSurv
	}
	if a.Penalty < b.Penalty {
		res.Penalty=a.Penalty
	} else {
		res.Penalty=b.Penalty
	}
	return res
}

func fitGen(IndCh chan<- NN.Individual, ResCh <-chan NN.Vector) EA.FitFunc {
	return func(ind EA.Individual) EA.ContextFit {
		var a, k, s, p float64
		for i := 0; i<5; i++ {
			IndCh<-*(ind.(*NN.Individual))
			res:=<-ResCh
			a+=res[0]
			k+=res[1]
			s+=res[2]
			p+=res[3]
		}
		a,k,s,p=a/5,k/5,s/5,p/5
		ft := 0.5*a+0.4*k+0.1*s/(60e9)-0.01*p//formula
		return ContextFit{ft,a,k,s,p}
	}
}

func genUniqueKey(alg EA.EA, nn NN.NN) string {
	return fmt.Sprintf("EA[%s]NN[%s]",alg.GetType(),nn.GetType())
}
