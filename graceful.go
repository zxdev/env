package env

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

/*

	grace := env.NewGraceful().Silent()
	...
	grace.Manager(&something)
	grace.Done() // wait on manager completion
	grace.Wait() // wait on shutdown signal

*/

// graceful struct control elements
type graceful struct {
	wgInit, wgShutdown *sync.WaitGroup
	ctx                context.Context
	cancel             context.CancelFunc
	silent, frame      bool
	exit               int
	name               string
	stop, wait, bye    atomic.Bool
}

// NewGraceful configurator returns *graceful and starts the shutdown controller to
// capture (os.Interrupt, syscall.SIGTERM, syscall.SIGHUP) signals and waits on
// the <-g.context.Done() for a signal and waits for the g.Manager
// controller wgShutdown to confirm all managed processes and completed tasks before
// the program terminates execution
func NewGraceful() *graceful {

	g := &graceful{
		wgInit:     new(sync.WaitGroup),
		wgShutdown: new(sync.WaitGroup),
		name:       filepath.Base(os.Args[0]),
	}
	g.ctx, g.cancel = context.WithCancel(context.Background())

	go func(g *graceful) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		select {
		case <-g.ctx.Done(): // program flow signal
		case <-sig: // system interrupt or user sighup|sigterm signal
			signal.Stop(sig) // got a signal; one is enough
			g.cancel()
		}
		g.Wait()
	}(g)

	return g
}

// Init starts a gracefully manged initilization func() or func(ctx,init)
//
//	a non-blocking func() exits and then the init.Done() triggers externally to signal completion; while
//	a blocking func(ctx,init) expectes init.Done() to trigger internally to signal completion and ready
//	state before internally blocking and waiting on a <-ctx.Done() to signal any cleanup actions on exit
//
// these signatures confirm the ready state of the started process via grace.Done()
//
//	func() { return }
//	func(context.Context, *sync.WaitGroup) { init.Done(); <-ctx.Done(); return }
//
// this signature can only confirm the process has started since the ready state is indeterminate via grace.Done()
//
//	func(context.Context)
func (g *graceful) Init(obj ...interface{}) *graceful {

	if g == nil {
		g = NewGraceful()
		log.Printf("grace: %s init [%d]", g.name, len(obj))
	}

	g.wgInit.Add(len(obj) + 1)
	defer g.wgInit.Done()

	for i := range obj {

		g.wgShutdown.Add(1)
		go func(obj interface{}, wgInit *sync.WaitGroup) {
			defer g.wgShutdown.Done()
			switch fxn := obj.(type) {
			// func() expected to be non-blocking and wgInit.Done()
			// tiggers afer the function returns; call to grace.Done()
			// will confirm ready state
			case func():
				fxn()
				wgInit.Done()
			// func(context.Context, *sync.WaitGroup) expected to block
			// and wgInit.Done() triggers before context blocking occurs;
			// call to grace.Done() confirms ready state
			case func(context.Context, *sync.WaitGroup):
				fxn(g.Context(), wgInit)
			// func(context.Context) legacy blocks on context, but can only signal
			// the process has started; a call to grace.Done() can not confirm the
			// ready state, it can only signal GraceInit started the process
			case func(context.Context):
				wgInit.Done()
				fxn(g.Context())
			}

		}(obj[i], g.wgInit)
		time.Sleep(time.Millisecond) // go routine ordering control

	}

	return g
}

// Silent log flag toggle that writes logs on os.Stderr (default: on)
func (g *graceful) Silent() *graceful { g.silent = !g.silent; return g }

// Frame bar flag toggle for graceful events (default:on)
func (g *graceful) Frame() *graceful { g.silent = !g.silent; return g }

// SetExit sets the os.Exit(n) status code
//
// zero causes a simple return instead of os.Exit
func (g *graceful) SetExit(i int) *graceful { g.exit = i; return g }

// Context is the background master g.context expored for use where the
// background context from graceful should be extended to other processes
func (g *graceful) Context() context.Context { return g.ctx }

// Cancels the graceful background context and waits for a clean exit;
// is order flowed to abort multiple calls
func (g *graceful) Cancel() {
	if g.stop.CompareAndSwap(false, true) {
		// if !g.silent {
		// 	log.Printf("%s: shutdown initiated", g.name)
		// }
		g.framer("shutdown initiated")
		g.cancel() // signal manager shutdowns
		g.Wait()
	}
}

// Done blocks until all g.wgInit process are complete
func (g *graceful) Done() {
	// delay timer to allow g.Manager to register
	// at least one wgInit.Add(1) event
	time.Sleep(time.Millisecond * 250)
	g.wgInit.Wait()
	//if !g.silent {
	g.framer("initilization complete")
	//log.Printf("%s: initilization complete", g.name)
	//}
}

// Wait blocks on the g context and waits for inits to terminate to cleanly exit
// when a g.exit value is non-zero the process will call os.Exit(n), otherwise
// exits with a simple return and is order flow controlled to abort multiple calls
func (g *graceful) Wait() {
	if g.wait.CompareAndSwap(false, true) { // ignore recurrent calls

		g.wgInit.Wait()     // allow init bootstraps to complete
		<-g.ctx.Done()      // block and wait on context
		g.wgShutdown.Wait() // allow shutdowns to complete

		if g.bye.CompareAndSwap(false, true) { // ignore recurrent calls
			g.framer("bye")
			// if !g.silent {
			// 	log.Printf("|%s|", strings.Repeat("-", 40))
			// 	log.Printf(" %s: bye", g.name)
			// 	log.Printf("|%s|", strings.Repeat("-", 40))
			// }
			time.Sleep(time.Millisecond * 250)
			if g.exit != 0 {
				os.Exit(g.exit)
			}
		}
	}
}

// framer is the bar frame content printer
func (g *graceful) framer(event string) {
	if !g.silent {
		if !g.frame {
			log.Printf("|%s|", strings.Repeat("-", 40))
		}
		log.Printf(" %s: %s", g.name, event)
		if !g.frame {
			log.Printf("|%s|", strings.Repeat("-", 40))
		}
	}
}
