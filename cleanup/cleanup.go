package cleanup

import (
	"os"
	"os/signal"
	"syscall"
	"sync"
)

var wg sync.WaitGroup

type trapFn func() 

func Trap(fn trapFn) trapFn {
	wg.Add(1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	d := make(chan struct{}, 1)

	ret := func() {
		fn()
		wg.Done()
		close(d)
	}

    go func() {
		defer func() {
			wg.Wait()
			if c != nil {
				os.Exit(1)
			}
		}()

		select {
			case <-c:
				ret()
				break
			case <-d:
				c = nil
				break
		}
	}()


	return ret
}
