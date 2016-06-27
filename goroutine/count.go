package goroutine

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func init() {
	go func() {
		tick := time.NewTicker(time.Second * 1)
		for {
			select {
			case <-tick.C:
				fmt.Fprintf(os.Stderr, "goroutine count: %d\n", runtime.NumGoroutine())
			}
		}
	}()
}
