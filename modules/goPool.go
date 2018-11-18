package modules

type goRoutinePool struct {
	active bool
	ch     chan byte
}

func newGoPool(cap int) *goRoutinePool {
	goPool := &goRoutinePool{
		ch: make(chan byte, cap),
	}
	for i := 0; i < cap; i++ {
		goPool.ch <- 1
	}
	goPool.active = true
	return goPool
}

// Take 取
func (p *goRoutinePool) Take() {
	<-p.ch
}

// Return 还
func (p *goRoutinePool) Return() {
	p.ch <- 1
}

// Close 关
func (p *goRoutinePool) Close() {
	close(p.ch)
}
