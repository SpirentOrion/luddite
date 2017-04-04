package luddite

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/SpirentOrion/logrus"
)

func dumpGoroutineStacks(logger *log.Logger) {
	sigs := make(chan os.Signal, 1)
	go func() {
		for {
			<-sigs
			buf := make([]byte, 1<<16)
			size := runtime.Stack(buf, true)
			logger.Infof("*** goroutine dump ***\n%s", buf[:size])
		}
	}()
	signal.Notify(sigs, syscall.SIGUSR1)
}
