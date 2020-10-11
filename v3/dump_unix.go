// +build !windows

package luddite

import (
	"os"
	"os/signal"
	"runtime"
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

func goroutineStack(all bool) []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, all)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
