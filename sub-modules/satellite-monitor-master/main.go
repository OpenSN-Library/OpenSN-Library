package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"satellite/monitor/data"
	"satellite/monitor/utils"
	"strconv"
)

func main() {
	args := os.Args
	var portStr string
	var servicePort string
	if len(args) == 1 {
		portStr = os.Getenv("UDP_PORT")
		servicePort = os.Getenv("SERVICE_PORT")
	} else if len(args) == 2 {
		panic("Usage: ./monitor [port] [service port]")
	} else if len(args) == 3 {
		portStr = args[1]
		servicePort = args[2]
	} else {
		panic("Usage: ./monitor [port] [service port]")
	}
	if servicePort == "" {
		servicePort = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	data.InitData()
	go utils.RecvUpdate("0.0.0.0", port)
	go utils.RecvConn("0.0.0.0", port+1)
	r := gin.Default()
	register(r)
	err = r.Run("0.0.0.0:" + servicePort)
	if err != nil {
		panic(err)
	}
}
