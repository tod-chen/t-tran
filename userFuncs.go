package main

import (
	"time"
	"12306/modules"
)

const(
	// queryTimeout 用户发起请求的超时时间	
	queryTimeout time.Duration = 5 * time.Second
)

// queryTranResult 查询车次及余票结果
type queryTranResult struct{
	depStations map[string]string
	arrStations map[string]string
	tranInfos []string
}

// 查询车次及余票数
func queryTrans(depStation, arrStation string, depDate time.Time, isAdult bool)queryTranResult{
	_, depCityCode := modules.GetCityCodeByStationName(depStation)
	_, arrCityCode := modules.GetCityCodeByStationName(arrStation)
	result := queryTranResult{
		// depStations : modules.GetRelationStations(depCityCode),
		// arrStations : modules.GetRelationStations(arrCityCode),
	}
	result.tranInfos = modules.QueryMatchTransInfo(depCityCode, arrCityCode, depDate, isAdult)
	return result
}

func queryRoute(tranNum string)[]modules.QueryRouteResult {
	return modules.QueryRoutetable(tranNum)
}

// 提交订单
func submitOrder(){

}

// 查询订单
func queryOrder(){

}

// 取消订单
func cancelOrder(){

}

// 支付订单
func paymentOrder(){

}

// 退票
func refundOrder(){

}

// 改签
func changeOrder(){
	
}

// 取票
func printTicket(){

}


// func queryTrans(departureStation, arrivalStation string, date time.Time, isAdult bool) []queryResult{
// 	var result []queryResult
// 	trans := []tran{}	// 所有车次
// 	deps := strings.Split(departureStation, ",")
// 	arrivals := strings.Split(arrivalStation , ",")
// 	var wg sync.WaitGroup 
// 	var l sync.Mutex
// 	for _, item := range trans{
// 		if !isMatchTranType(&item, types){
// 			continue
// 		}
// 		if depIndex, arrIndex, ok := isMatchStationAndTime(&item, deps, arrivals, date, departureStartTime, departureEndTime); ok{
// 			wg.Add(1)
// 			go func (t *tran, depIndex, arrIndex int){
// 				defer wg.Done()
// 				var data queryResult
// 				data.tranNum = t.tranNum
// 				data.tranType = t.tranType
// 				data.departureTime = t.routeTimetable[depIndex].departureTime
// 				data.arrivalTime = t.routeTimetable[arrIndex].arrivalTime
// 				l.Lock()
// 				result = append(result, data)
// 				l.Unlock()
// 			}(&item, depIndex, arrIndex)
// 		}
// 	}
// 	wg.Wait()
// 	return result
// }

// func isMatchTranType(t *tran, types []string) bool{
// 	for _, item := range types{
// 		if t.tranType == item{
// 			return true
// 		}
// 	}
// 	return false
// }


// func book(tranID, userID, contactID, depIndex, arrIndex int, isAdult bool, seatType string)(string, bool){
// 	var t tran // 根据tranID获取车次
// 	msg, ok := t.Book(tranID, depIndex, arrIndex, userID, contactID, isAdult, seatType)
// 	return msg, ok
// }

// func pay(userID, payType int, payAccount string, price float32){

// }