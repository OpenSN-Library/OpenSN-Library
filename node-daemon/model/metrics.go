package model

import (
	"NodeDaemon/utils"
	"time"
)

type HostResource struct {
	Time time.Time
	utils.HostResource
}

type InstanceResouce struct {
	Time time.Time
	utils.InstanceResouce
}

type LinkResource struct {
	Time time.Time
	utils.LinkResource
}