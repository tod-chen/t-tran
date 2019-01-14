package modules

import "sync"

// IDPool ID池
type IDPool struct {
	capacity int
	len      int
	pool     chan uint64
	key      string
	sync.Mutex
}

func (p *IDPool) pull() {

	var offset uint64
	db.Table("config_id").Where("key = ?", p.key).Pluck("val", &offset)
	// TODO：确认更新语句成功
	db.Exec("update config_id set val = ? where key = ? and val = ?", offset+uint64(p.capacity), p.key, offset)
	for i := 0; i < p.capacity; i++ {
		p.pool <- (offset + uint64(i))
	}
}

func (p *IDPool) getID(baseID uint64) uint64 {
	p.Lock()
	id := <-p.pool
	m := baseID % 10 // 模10
	id = (id << 4) ^ m
	p.len--
	if p.len == 0 {
		p.pull()
	}
	p.Unlock()
	return id
}

var (
	orderIDPool  IDPool
	ticketIDPool IDPool
)

func initGenerateID() {
	cap := 10000
	orderIDPool = IDPool{
		capacity: cap,
		len:      cap,
		pool:     make(chan uint64, cap),
		key:      "order_id",
	}
	orderIDPool.pull()

	ticketIDPool = IDPool{
		capacity: cap,
		len:      cap,
		pool:     make(chan uint64, cap),
		key:      "ticket_id",
	}
	ticketIDPool.pull()
}

func getOrderID(userID uint64) uint64 {
	return orderIDPool.getID(userID)
}

func getTicketID(passengerID uint64) uint64 {
	return ticketIDPool.getID(passengerID)
}

func getMultiTicketID(passengerIDs []uint64) []uint64 {
	pLen := len(passengerIDs)
	result := make([]uint64, pLen)
	for i := 0; i < pLen; i++ {
		id := ticketIDPool.getID(passengerIDs[i])
		result[i] = id
	}
	return result
}
