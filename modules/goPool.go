package modules

import(
	"fmt"
)

const (
	constMaxGoroutineCount = 5000
)

// goroutine 池
var goPool *goRoutinePool

func initGoPool(){
	fmt.Println("begin init goPool")
	defer fmt.Println("end init goPool")
	goPool = &goRoutinePool{}
	goPool.ch = make(chan byte, constMaxGoroutineCount)
	for i:=0;i<constMaxGoroutineCount;i++{
		goPool.ch <- 1
	}
	goPool.active = true
}

type goRoutinePool struct{
	active bool
	ch chan byte
}

func (p *goRoutinePool)Take(){
	<- p.ch
}

func (p *goRoutinePool)Return(){
	p.ch <- 1
}