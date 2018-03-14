package modules

const (
	maxGoroutineCount = 5000
)

var goPool *goRoutinePool

func init(){
	goPool = &goRoutinePool{}
	goPool.ch = make(chan byte, maxGoroutineCount)
	for i:=0;i<maxGoroutineCount;i++{
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