package model

type WebShellAllocRequest struct {
	WebShellID   string   `json:"webshell_id"`
	Command      string   `json:"command"`
	Args         []string `json:"args"`
	Writeable    bool     `json:"writeable"`
	ExpireMinute int    `json:"expire_minute"`
}

type WebShellAllocInfo struct {
	WebShellID string `json:"webshell_id"`
	Addr       string `json:"addr"`
	Port       int    `json:"port"`
	Pid        int    `json:"pid"`
}
