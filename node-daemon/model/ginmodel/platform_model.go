package ginmodel

type RedisConfiguration struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Index    int    `json:"index"`
}

type EtcdConfiguration struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}