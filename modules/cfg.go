package modules

type config struct {
	Key   string `bson:"key"`   // 配置名
	Value string `bson:"value"` // 配置值
}
