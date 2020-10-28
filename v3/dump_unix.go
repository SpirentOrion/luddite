// +build !windows

package luddite

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var dumpOnce sync.Once

func dumpGoroutineStacks() {
	dumpOnce.Do(func() {
		sigs := make(chan os.Signal, 1)
		go func() {
			for {
				<-sigs
				os.Stderr.Write(goroutineStack(true))
				os.Stderr.Write([]byte("\n"))
			}
		}()
		signal.Notify(sigs, syscall.SIGUSR1)
	})
}
