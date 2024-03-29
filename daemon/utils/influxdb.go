package utils

import (
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var InfluxDBClient influxdb2.Client
var InfluxDBWriteAPI api.WriteAPIBlocking
var InfluxDBQueryAPI api.QueryAPI

func InitInfluxDB(addr, token, org, bucket string, port int) error {
	client := influxdb2.NewClient(
		fmt.Sprintf(
			"http://%s:%d",
			addr,
			port,
		),
		token,
	)
	InfluxDBClient = client
	InfluxDBWriteAPI = InfluxDBClient.WriteAPIBlocking(org, bucket)
	InfluxDBQueryAPI = InfluxDBClient.QueryAPI(org)
	return nil
}
