package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	sigChann := make(chan os.Signal,1)
	signal.Notify(sigChann,syscall.SIGTERM)
	go func ()  {
		<- sigChann
		wg.Done()	
	}()
	wg.Wait()
}
