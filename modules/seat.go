package modules

import (
	"sync"
)


type seat struct{
	l sync.RWMutex
	tranID int
	carNum int
	seatNum string
	seatType string
	isAdult bool
	isFull bool
	isPutTogetherNoSeat bool
	seatRoute uint32
}

func (s *seat)IsAvailable(seatMatch uint32, isAdult bool) bool{
	if s.isAdult != isAdult || s.isFull{
		return false
	}
	return s.seatRoute ^ seatMatch == s.seatRoute + seatMatch
}

func getSeatMatch(depIndex, arrIndex uint) (result uint32) {
	for i:=depIndex; i<= arrIndex; i++ {
		result ^= 1 << i
	}
	return
}

var allSeatPrices []seatPrice
type seatPrice struct{
	tranNum string
	seatTypeIndex uint
	priceEachRoute []float32
}

func getTicketPrice(tranNum string, typeIndex, depIndex, arrIndex uint) float32{
	var price float32
	for _, item := range allSeatPrices{
		if item.tranNum == tranNum && item.seatTypeIndex == typeIndex {
			for i:=depIndex; i<arrIndex; i++ {
				price += item.priceEachRoute[i]
			}
		}
	}
	return price
}

