// MIT License

// Copyright (c) 2020 zxdev

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package env

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"time"
)

/*

	// var s Sample
	// env.Manage(&s)
	//  .. or ..
	// s := new(Sample)
	// env.Manage(s)

	type Sample struct{}
	func (s *Sample) Start() env.GracefulFunc {
		// startup jobs and definatios here
		return func(ctx context.Context) {
			<-ctx.Done()
			// shutdown and cleanup here
		}
	}

	// env.Manage(Example)
	func Example(ctx context.Context) env.GracefulFunc {
		// startup jobs and definatios here
		return func(ctx context.Context) {
			<-ctx.Done()
			// shutdown and cleanup here
		}
	}

	// env.Manage(Param(42))
	func Param(n int) env.GracefulFunc {
		go func(){ // prevent lock-step with go routine wrapper
		// startup jobs and definatios here
		}()
		return func(ctx context.Context) {
			<-ctx.Done()
			// shutdown and cleanup here
		}
	}

*/

var (
	// ctx and wg for graceful management
	ctx, cancel  = context.WithCancel(context.Background())
	wgInitialize = new(sync.WaitGroup) // initialize control group
	wgShutdown   = new(sync.WaitGroup) // shutdown control group
)

// Context returns the master graceful context.Context
func Context() context.Context { return ctx }

// GracefulFunc controller type
type GracefulFunc func(ctx context.Context)

// Graceful controller interface type
type Graceful interface {
	Start() GracefulFunc
}

// run manager; does not enter here until passed in function
// has completely executed and returned the shutdown function
func run(shutdown GracefulFunc, name string) {

	wgInitialize.Done()
	shutdown(ctx)
	if summary {
		log.Printf("%s: stop", name)
	}
	wgShutdown.Done()
}

// Manage start/stop gracefully requires Graceful interface signature or a graceful
// function with with a graceful signature following func(ctx context.Context) format;
// the optional name is automatically extracted from Graceful interface types or is
// randomly generated when a name is not supplied with a graceful function
func Manage(g interface{}, name ...string) {

	wgInitialize.Add(1)
	wgShutdown.Add(1)

	switch g.(type) {
	case Graceful:
		if len(name) == 0 { // extract name reference
			if reflect.TypeOf(g).Kind() == reflect.Ptr {
				name = []string{strings.ToLower(reflect.TypeOf(g).Elem().Name())}
			} else {
				name = []string{strings.ToLower(reflect.TypeOf(g).Name())}
			}
		}

	default:
		if len(name) == 0 {
			var b [4]byte
			rand.Read(b[:])
			name = []string{fmt.Sprintf("%x", b)}
		}
	}

	if summary {
		log.Printf("%s: start", name[0])
	}

	switch g.(type) {
	case Graceful:
		// func (e *Example) Start() env.GracefulFunc {
		//  return func(ctx context.Context) {
		//   <-ctx.Done()
		//  }
		// }
		// env.Manage(&cfg)
		go func() { run(g.(Graceful).Start(), name[0]) }()

	case func() GracefulFunc:
		// func sample() env.GracefulFunc {
		//  return func(ctx context.Context) {
		//   <-ctx.Done()
		//  }
		// }
		// env.Manage(sample, "sample")
		go func() { run(g.(func() GracefulFunc)(), name[0]) }()

	case GracefulFunc:
		// func (ex *Example) Connect(param string) env.GracefulFunc {
		//  go func(){ // wrap to avoid lock-step blocking start
		//  }()
		//  return func(ctx context.Context) {
		//   <-ctx.Done()
		//  }
		// }
		// env.Manage(ex.Connect("param"), "connect")
		go func() { run(g.(GracefulFunc), name[0]) }()

	case func(ctx context.Context):
		// func sample(ctx context.Context) {
		//   <-ctx.Done()
		//  }
		// env.Manage(func(),"sample")
		go func() { run(g.(func(ctx context.Context)), name[0]) }()

	default:
		fmt.Fprintln(os.Stderr, "env: unrecognized graceful interface")
		os.Exit(1)
	}

}

// Ready blocks until all Initializations have completed.
func Ready() {

	// avoid panic: sync: WaitGroup is reused before previous Wait has returned
	time.Sleep(time.Millisecond * 100)

	wgInitialize.Wait()
	if summary {
		messageBar("log")
	}

}

// Stop immediately signals all graceful controllers to begin the
// shutdown sequence, blocking until all have existed.
//
// Stop() should only be called as the last item in a main() process.
func Stop() {

	wgInitialize.Wait() // ensure all initializations completed
	if summary {
		messageBar("shutdown")
	}

	cancel()          // signals graceful controllers to exit
	wgShutdown.Wait() // wait until all graceful controllers exit
	if summary {
		time.Sleep(time.Millisecond * 100) // delay allows log of shutdown bye message
	}

}

var shutdown bool

// Shutdown waits for a termination signal, initiates graceful cleanup, then calls os.Exit;
// Stop() or an os.Interrupt or os.Kill signal is trigger event; configure only once
func Shutdown() {
	if !shutdown {

		// use case examples; blocking
		//  defer Shutdown()
		//  Shutdown()
		//
		// or non-blocking
		//  go Shutdown()

		shutdown = true // mark as initialized
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)

		for {
			select {
			case <-ctx.Done():
				// wait until all graceful controllers exit
				wgShutdown.Wait()
				if summary {
					messageBar("bye")
				}
				os.Exit(0)

			case s := <-sig:
				signal.Stop(sig)
				// signal graceful controllers to exit
				cancel()
				if summary {
					messageBar(s.String())
				}
			}
		}

	}
}
