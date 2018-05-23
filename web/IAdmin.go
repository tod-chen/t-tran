package web

// IAdmin 管理员行为接口
type IAdmin interface{
	// 添加车次信息
	addTranInfo()
	// 修改车次信息
	editTranInfo()
	// 删除车次信息
	deleteTranInfo()
	// 修改车次的时刻表
	editRoute()
	// 修改车次的座次类型及数量	
	editSeat()
}