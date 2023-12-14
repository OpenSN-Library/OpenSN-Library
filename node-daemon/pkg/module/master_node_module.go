package biz

import (
	"NodeDaemon/config"
	"fmt"
	"net/http"
	"sync"
)

const MasterNodeContainerName = "master_node"

type MasterNodeModule struct {
	ModuleBase
	containerID string
}

func (m *MasterNodeModule) IsSetupFinish() bool {
	url := fmt.Sprintf("http://%s:8080/api/platform/status", config.MasterAddress)
	_, err := http.Get(url)
	return err == nil
}

func masterDaemonFunc(sigChann chan int, errChann chan error) {
	
}

func CreateMasterNodeModuleTask() *MasterNodeModule {
	return &MasterNodeModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: masterDaemonFunc,
			wg:         new(sync.WaitGroup),
		}, "",
	}
}
