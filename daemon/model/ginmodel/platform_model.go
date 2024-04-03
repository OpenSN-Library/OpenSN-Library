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

type InfluxDBConfiguration struct {
	Enable  bool   `json:"enable"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Org     string `json:"org"`
	Bucket  string `json:"bucket"`
	Token   string `json:"token"`
}

type CodeServerConfiguration struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}
