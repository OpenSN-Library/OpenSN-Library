package biz

import (
	"NodeDaemon/share/signal"
	"sync"
	"time"
)

var ModuleCheckGap = 1 * time.Minute

type Module interface {
	Run()
	Stop()
	CheckError() error
	IsRunning() bool
	Wait()
}

type ModuleBase struct {
	sigChan    chan int
	errChan    chan error
	runing     bool
	daemonFunc func(sigChann chan int, errChann chan error)
	wg         *sync.WaitGroup
}

func (m *ModuleBase) Run() {
	if m.IsRunning() {
		return
	}
	m.runing = true
	m.wg.Add(1)
	go func() {
		m.daemonFunc(m.sigChan, m.errChan)
		m.runing = false
		m.wg.Done()
	}()
}

func (m *ModuleBase) Stop() {
	m.sigChan <- signal.STOP_SIGNAL
}

func (m *ModuleBase) CheckError() error {
	select {
	case res := <-m.errChan:
		m.errChan <- res
		return res
	default:
		return nil
	}
}

func (m *ModuleBase) IsRunning() bool {
	return m.runing
}

func (m *ModuleBase) Wait() {
	m.wg.Wait()
}
