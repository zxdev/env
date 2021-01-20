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
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	// ctx and wg for graceful management
	ctx, cancel = context.WithCancel(context.Background())
	wgManager   = new(sync.WaitGroup) // control group for Manage()
	wgShutdown  = new(sync.WaitGroup) // shutdown control group
)

// Context returns the master graceful context.Context
func Context() context.Context { return ctx }

// gracefulStruct controller interface type
type gracefulStruct1 interface {
	Start() func(ctx context.Context)
}

type gracefulStruct2 interface {
	Start() func(ctx context.Context) // start method
	Name() string                     // optional, alternate name other than
}

// gracefulStruct3 controller interface type
type gracefulStruct3 interface {
	Start(ctx context.Context)
}

// gracefulStruct4 controller interface type
type gracefulStruct4 interface {
	Start(ctx context.Context) // start method
	Name() string              // optional, alternate name other than
}

// Manager places the struct or func under graceful management and references it by name
// and requires the function sigature of 'func(ctx context.Context)', so any closure must
// return the graceful signature to operate properly as well.
//
// A graceful struct must have a 'Start()' method with the proper graceful signature, and
// to override the use the struct name as defined in the code a struct my define an
// optional 'Name() string' signature method to provide an alternative name reference.
//
// The use of any params passed to a struct or function that uses a closure will make the
// sequence run lock-step. To aloid this, wrap the closure head in a go routine.
func Manager(g interface{}) {

	wgManager.Add(1)
	wgShutdown.Add(1)

	switch {
	case reflect.TypeOf(g).Kind() == reflect.Ptr &&
		reflect.TypeOf(g).Elem().Kind() == reflect.Struct:

		switch g.(type) {

		case gracefulStruct4: // Start(ctx context.Context); Name() string
			go func() {
				if summary {
					log.Printf("%s: start", strings.ToLower(g.(gracefulStruct4).Name()))
					defer log.Printf("%s: stop", strings.ToLower(g.(gracefulStruct4).Name()))
				}
				wgManager.Done()
				g.(gracefulStruct4).Start(ctx)
				wgShutdown.Done()
			}()

		case gracefulStruct3: // Start(ctx context.Context)
			go func() {
				if summary {
					log.Printf("%s: start", strings.ToLower(reflect.TypeOf(g).Elem().Name()))
					defer log.Printf("%s: stop", strings.ToLower(reflect.TypeOf(g).Elem().Name()))
				}
				wgManager.Done()
				g.(gracefulStruct3).Start(ctx)
				wgShutdown.Done()
			}()

		case gracefulStruct2: // Start() func(ctx context.Context); Name() string
			go func() {
				if summary {
					log.Printf("%s: start", strings.ToLower(g.(gracefulStruct2).Name()))
					defer log.Printf("%s: stop", strings.ToLower(g.(gracefulStruct2).Name()))
				}
				wgManager.Done()
				g.(gracefulStruct2).Start()(ctx)
				wgShutdown.Done()
			}()

		case gracefulStruct1: // Start() func(ctx context.Context)
			go func() {
				if summary {
					log.Printf("%s: start", strings.ToLower(reflect.TypeOf(g).Elem().Name()))
					defer log.Printf("%s: stop", strings.ToLower(reflect.TypeOf(g).Elem().Name()))
				}
				wgManager.Done()
				g.(gracefulStruct1).Start()(ctx)
				wgShutdown.Done()
			}()

		default:
			log.Println("alert: unsupported struct type")
			os.Exit(0)
		}

	case reflect.TypeOf(g).Kind() == reflect.Func:

		// struct:  convert insite/pkg/env_test.(*Gamma).Basic-fm to 'basic'
		// package: insite/pkg/server.Start.func3 to 'server'
		name := filepath.Base(runtime.FuncForPC(reflect.ValueOf(g).Pointer()).Name()) // identify func by name
		if strings.Contains(name, "*") {
			name = strings.Split(name, ".")[strings.Count(name, ".")] // extract last segment
			name = strings.TrimSuffix(name, "-fm")                    // clean up tail
		} else {
			name = strings.Split(name, ".")[0]
		}
		name = strings.ToLower(name)

		switch g.(type) {
		case func() func(ctx context.Context): // func() func(ctx context.Context)
			go func() {
				if summary {
					log.Printf("%s: start", name)
					defer log.Printf("%s: stop", name)
				}
				wgManager.Done()
				g.(func() func(ctx context.Context))()(ctx)
				wgShutdown.Done()
			}()

		case func(ctx context.Context): // func(ctx context.Context)
			go func() {
				if summary {
					log.Printf("%s: start", name)
					defer log.Printf("%s: stop", name)
				}
				wgManager.Done()
				g.(func(ctx context.Context))(ctx)
				wgShutdown.Done()
			}()

		default:
			log.Printf("alert: %s unsupported func type", name)
			os.Exit(0)
		}

	default:
		log.Println("alert: unsupported type")
		os.Exit(0)
	}

}

// Ready blocks until all Initializations have completed.
func Ready() {

	// avoid panic: sync: WaitGroup is reused before previous Wait has returned
	time.Sleep(time.Millisecond * 100)

	wgManager.Wait()
	if summary {
		messageBar("log")
	}

}

// Stop immediately signals all graceful controllers to begin the
// shutdown sequence, blocking until all have existed.
//
// Stop() should only be called as the last item in a main() process.
func Stop() {

	wgManager.Wait() // ensure all initializations completed
	if summary {
		messageBar("shutdown")
	}

	cancel() // signal graceful controllers to exit
	if summary {
		defer messageBar("bye")
	}
	wgShutdown.Wait() // wait until all graceful controllers exit

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
