package main

import (
	"fmt"
	"AI/NN"
	"time"
	"math/rand"
	"math"
	"path/filepath"
	"os"
	"encoding/json"
	"github.com/veandco/go-sdl2/sdl"
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

const iters int = 1

const cellSize int = 8

type ContextFit struct {
	Fit float64
	ApplesEaten float64
	Kills float64
	TimeSurv float64//dur in nanosec
	Penalty float64
}

type BestIndividualJSON struct {
	Type string
	Layers []int
	Ind NN.Individual
	//Fit ContextFit
}

func Color(alg, nn string) (uint8,uint8,uint8) {
	var x, y uint8
	switch alg {
		case "GA":x=0;y=0
		case "PSO":x=0;y=127
		case "ABC":x=127;y=0
		case "SSO":x=127;y=127
		default:x=50;y=50
	}
	switch nn {
		case "FNN":return 255,x,y
		case "RNN":return x,255,y
		case "LSTM":return x,y,255
		default:return 255,255,255
	}
}

func main() {
	rand.Seed(int64(time.Now().Nanosecond()))
	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()
	NNs := []NN.NN{NN.FNNtempl,NN.RNNtempl, NN.LSTMtempl}
	EAs := []string{"GA", "PSO", "ABC", "SSO"}
	window, err := sdl.CreateWindow("Snake", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,int32(100*cellSize), int32(100*cellSize), sdl.WINDOW_SHOWN)
	defer window.Destroy()
	if err != nil {fmt.Printf("Failed to create window: %s\n", err); os.Exit(1)}
	rend, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	defer rend.Destroy()
	if err != nil {fmt.Printf("Failed to create renderer: %s\n", err); os.Exit(1)}
	rend.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	Clear(rend, 0, 0, 0, 255)
	rend.Present()
	for _, nnt := range NNs {
		window.Show()
		window.Raise()
		nn:=nnt.GetType()
		for _, ea := range EAs {
			window.Show()
			window.Raise()
			key:=fmt.Sprintf("EA[%s]NN[%s]",ea,nn)
			r, g, b := Color(ea,nn)
			f, _ := os.Open(filepath.Join(".","Trained",key+".final.json"))
			st,_:=f.Stat()
			data := make([]byte, st.Size())
			f.Read(data)
			f.Close()
			var res BestIndividualJSON
			json.Unmarshal(data, &res)
			for i := 0; i < iters; i++ {
				GlobalDB.Snakes = make(map[string]SnakeData, 0)
				GlobalDB.Apples = make([]Coord, 0)
				GlobalDB.MapSize = Coord{100, 100}
				GlobalDB.mutex = make(chan bool, 1)
				GlobalDB.mutex <- true
				nn,_:=NN.New(nnt,res.Layers...)
				nn.GenFromIndividual(res.Ind)
				ResChan := make(chan string, 1)
				var CommChansS map[string](chan bool) = map[string](chan bool){key:make(chan bool,1),"player":make(chan bool,1)}
				var CommChansR map[string](chan CommandLine) = map[string](chan CommandLine){key:make(chan CommandLine,1),"player":make(chan CommandLine,1)}
				EventChan := make(chan *sdl.KeyboardEvent,1)
				stop := false
				fmt.Println("Playing with:", key)
				go gameAgent(key,nn,CommChansS[key],CommChansR[key])
				go playerHandler(rend, CommChansS["player"],CommChansR["player"], EventChan, &stop, r, g, b)
				go gameHandler(&CommChansS,&CommChansR,ResChan)
				for !stop {
					event:=sdl.WaitEventTimeout(100)
					switch t:=event.(type) {
						case *sdl.KeyboardEvent:
							EventChan<-t
					}
				}
				res := <-ResChan
				fmt.Println("Died:", res)
				Clear(rend, 0, 0, 0, 255)
				rend.Present()
			}
		}
	}
}

func playerHandler(rend *sdl.Renderer, gameChanReceive <-chan bool, gameChanSend chan<- CommandLine, EventChan chan *sdl.KeyboardEvent, stop *bool, r, g, b uint8) {
	gameChanSend<-CommandLine{register, "player", Coord{}}
	<-gameChanReceive
	gameChanSend<-CommandLine{spawn, "player", Coord{}}
	cmdSend := false
	for {
		cont:=<-gameChanReceive
		if cont {
			select {
				case t:=<-EventChan:
					if len(gameChanSend)<1 {
						if t.State==sdl.PRESSED {
							switch t.Keysym.Sym {
								case sdl.K_LEFT:
									gameChanSend<-CommandLine{do, "player", Coord{-1,0}}
									cmdSend=true
								case sdl.K_RIGHT:
									gameChanSend<-CommandLine{do, "player", Coord{1,0}}
									cmdSend=true
								case sdl.K_UP:
									gameChanSend<-CommandLine{do, "player", Coord{0,-1}}
									cmdSend=true
								case sdl.K_DOWN:
									gameChanSend<-CommandLine{do, "player", Coord{0,1}}
									cmdSend=true
							}
						}
					}
				default:
			}
			DataArr, self :=GlobalDB.GetDataWithMutex("player")
			Clear(rend, 0,0,0,255)
			DrawArr(rend, DataArr[1],255,0,0,255)//apples
			DrawArr(rend, DataArr[3],AddSomeX(r,50),AddSomeX(g,50),AddSomeX(b,50),255)//enemy body
			DrawArr(rend, DataArr[0],50,127,50,255)//player body
			SetCell(rend, self.Head.x, self.Head.y, 50,255,50,255)//player head
			DrawArr(rend, DataArr[2],r,g,b,255)//enemy heads
			rend.Present()
			if !cmdSend && len(gameChanSend)<1 {
				gameChanSend<-CommandLine{acknowledge, "player", Coord{}}
			}
			cmdSend=false
		} else {
			*stop=true
			break
		}
	}
}

func AddSomeX(a uint8, x uint8) uint8 {
	if a<255-x {
		return a+x
	} else {
		return a
	}
}

func Clear(renderer *sdl.Renderer, r, g, b, a uint8) {
	renderer.SetDrawColor(r, g, b, a)
	renderer.Clear()
}

func SetCell(renderer *sdl.Renderer, x, y int, r, g, b, a uint8) {
	renderer.SetDrawColor(r, g, b, a)
	renderer.FillRect(&sdl.Rect{int32(x*cellSize), int32(y*cellSize) , int32(cellSize), int32(cellSize)})
}

func DrawArr(renderer *sdl.Renderer, arr []Coord, r, g, b, a uint8) {
	for i := range arr {
		SetCell(renderer, arr[i].x, arr[i].y, r, g, b, a)
	}
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

func (c Coord) IsIn(min Coord, max Coord) bool {
	return c.x>=min.x && c.x<=max.x && c.y>=min.y && c.y<=max.y
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

func gameHandler(tickSend *map[string](chan bool), cmdRec *map[string](chan CommandLine), sendRes chan<- string) {
MAIN:
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
					head:=Coord{rand.Intn(dbcopy.MapSize.x-20)+10, rand.Intn(dbcopy.MapSize.y-20)+10}
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
		for k, s := range dbcopy.Snakes {
			if s.Alive {
				if s.Head.x<0 || s.Head.x>=dbcopy.MapSize.x || s.Head.y<0 || s.Head.y>=dbcopy.MapSize.y {
					s.Alive=false
					dbcopy.Snakes[k]=s
					sendRes <- k
					break MAIN
				}
				for _, b := range s.Body {
					if s.Head==b {
						s.Alive=false
						dbcopy.Snakes[k]=s
						sendRes <- k
						break MAIN
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
							sendRes <- "both"
							break MAIN
						}
						for _, b := range s2.Body {
							if s.Head==b {
								s.Alive=false
								s2.Kills++
								dbcopy.Snakes[k]=s
								dbcopy.Snakes[k2]=s2
								sendRes <- k
								break MAIN
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
	for k := range *tickSend {
		(*tickSend)[k]<-false
	}
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

func SelectIndex(vals NN.Vector) int {
	max := vals[0]
	best:=0
	for i:=1; i<len(vals); i++ {
		if max<vals[i] {
			max=vals[i]
			best=i
		}
	}
	return best	
}

func gameAgent(key string, nn NN.NN, gameChanReceive <-chan bool, gameChanSend chan<- CommandLine) {
	db := &GlobalDB
	var inGame bool
	gameChanSend<-CommandLine{register, key, Coord{}}
	HugeVal:=float64(db.MapSize.x*db.MapSize.y+100000)
MAIN:
	for {
		select {
			case cont:=<-gameChanReceive:
				if cont {
					if !inGame {
							inGame=true
							gameChanSend<-CommandLine{spawn, key, Coord{}}
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
							gameChanSend<-CommandLine{do, key, dir}
						} else {
							break MAIN
						}
					}
				} else {
					break MAIN
				}
			default:
		}
	}
}
