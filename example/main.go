package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/zxdev/env/v2"
)

type Sample struct{}

func (s *Sample) Start(ctx context.Context, wg *sync.WaitGroup) {
	// bootstrap process
	wg.Done()    // bootstrap completed
	<-ctx.Done() // block signal
	// shutdown process
}

type params struct {
	Name  string `env:"N,require,order,environ" help:"a name to use"`
	Flag  bool   `default:"on" hidden:"[redacted]" help:"a flag setting"`
	small int    // not parsed or reported in Summary
}

func main() {

	var param params
	param.small++
	ev := env.NewEnv(&param)
	//ev := env.NewSilentEnv(&param)
	log.Println(ev.Srv)
	sam := new(Sample)

	grace := env.NewGraceful()
	//grace.Silent()
	grace.Manager(sam)
	grace.Done()

	if param.Flag {

		// loop with timeout to signal shutdown
		for {
			select {
			case <-grace.Context().Done():
			case <-time.After(time.Minute):
			}
			grace.Stop()
		}

	} else {

		// wait for a ^C to terminate
		grace.Wait()

	}
}
