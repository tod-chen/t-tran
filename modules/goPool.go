package modules

const (
	constMaxGoroutineCount = 150
)

func initGoPool(cap int) *goRoutinePool {
	goPool := &goRoutinePool{}
	goPool.ch = make(chan byte, cap)
	for i := 0; i < cap; i++ {
		goPool.ch <- 1
	}
	goPool.active = true
	return goPool
}

type goRoutinePool struct {
	active bool
	ch     chan byte
}

func (p *goRoutinePool) Take() {
	<-p.ch
}

func (p *goRoutinePool) Return() {
	p.ch <- 1
}

func (p *goRoutinePool) Close() {
	close(p.ch)
}
