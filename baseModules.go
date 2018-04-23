package main

import (
	// "sync/atomic"
	// "sync"
	// "time"
)




// // GetAvailableSeatCount 获取余票数
// func GetAvailableSeatCount(seats []seat, depIndex, arrIndex int, isAdult bool)int32{
// 	length := len(seats)
// 	if length == 0{
// 		return -1
// 	}
// 	var wg sync.WaitGroup
// 	var count int32
// 	for i:=0;i<length;i++{
// 		if count >= queryMaxSeatCount {
// 			break
// 		}
// 		wg.Add(1)
// 		go func(s *seat){
// 			defer wg.Done()
// 			if count < queryMaxSeatCount && s.IsAvailable(depIndex, arrIndex, isAdult){
// 				atomic.AddInt32(&count, 1)
// 			}
// 		}(&seats[i])
// 	}
// 	wg.Wait()
// 	return count
// }

// // GetNoSeatAvailableCount 获取无座的余票数
// func GetNoSeatAvailableCount(seats []seat, depIndex, arrIndex int)int32{
// 	count := len(seats)
// 	if count == 0{
// 		return -1
// 	}
// 	var wg sync.WaitGroup
// 	array := make([]int32, 0, arrIndex + 1 - depIndex)	
// 	for i:=0;i<count;i++ {
// 		wg.Add(1)
// 		go func(s *seat){
// 			defer wg.Done()
// 			for i:=depIndex;i<=arrIndex;i++{
// 				if s.statusEachRoute[i] != 1{
// 					atomic.AddInt32(&array[depIndex-i], 1)
// 				}
// 			}
// 		}(&seats[i])
// 	}
// 	wg.Wait()
// 	var maxVal int32
// 	for _, a := range array{
// 		if a > maxVal{
// 			maxVal = a
// 		}
// 	}
// 	return int32(count) - maxVal
// }


