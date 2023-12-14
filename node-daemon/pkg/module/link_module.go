package module

import (
	"NodeDaemon/share/signal"

	"sync"
)

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	Base
}

func linkDaemonFunc(sigChann chan int, errChann chan error) {
watchLoop:
	for {
		select {
		case sig := <-sigChann:
			if sig == signal.STOP_SIGNAL {
				break watchLoop
			}
		}
	}
}

func CreateLinkModuleTask() *LinkModule {
	return &LinkModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: linkDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "LinkManage",
		},
	}
}
