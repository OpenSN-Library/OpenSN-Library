package module

import (
	"NodeDaemon/share/signal"
	"github.com/sirupsen/logrus"
	"sync"
)

type Module interface {
	Run()
	Stop()
	CheckError() error
	IsRunning() bool
	Wait()
}

type Base struct {
	sigChan    chan int
	errChan    chan error
	running    bool
	daemonFunc func(sigChann chan int, errChann chan error)
	wg         *sync.WaitGroup
	ModuleName string
}

func (m *Base) Run() {
	if m.IsRunning() {
		return
	}
	logrus.Infof("Run %s Module.", m.ModuleName)
	m.running = true
	m.wg.Add(1)
	go func() {
		m.daemonFunc(m.sigChan, m.errChan)
		m.running = false
		logrus.Infof("%s Module Stop.", m.ModuleName)
		m.wg.Done()
	}()
}

func (m *Base) Stop() {
	m.sigChan <- signal.STOP_SIGNAL
}

func (m *Base) CheckError() error {
	select {
	case res := <-m.errChan:
		m.errChan <- res
		return res
	default:
		return nil
	}
}

func (m *Base) IsRunning() bool {
	return m.running
}

func (m *Base) Wait() {
	m.wg.Wait()
}
