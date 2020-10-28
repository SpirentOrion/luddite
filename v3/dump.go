package luddite

import "runtime"

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
