package rpc

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	GrpcServerAddr = "127.0.0.1:50051"
)

var FrrClient NorthboundClient

func init() {
	// Set up a connection to the server.

	conn, err := grpc.Dial(GrpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Errorf("Connect Grpc Error: %s", err.Error())
		panic(err)
	}

	cli := NewNorthboundClient(conn)
	FrrClient = cli
}
