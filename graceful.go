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
	grace.Wait() // wait on manager completion
	grace.Shutdown() // wait on shutdown signal

*/

// graceful struct control elements
type graceful struct {
	init, shutdown  *sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	silent, frame   bool
	exit            int
	name            string
	stop, wait, bye atomic.Bool
	register        []func()
}

// NewGraceful configurator returns *graceful and starts a shutdown controller to
// capture (os.Interrupt, syscall.SIGTERM, syscall.SIGHUP) signals and waits on
// the <-g.context for a termination signal and waits for the g.init, g.shutdown
// controller shutdown to confirm all managed processes have completed tasks before
// the program terminates execution
func NewGraceful() *graceful {

	g := &graceful{
		init:     new(sync.WaitGroup),
		shutdown: new(sync.WaitGroup),
		name:     filepath.Base(os.Args[0]),
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
		g.Shutdown()
	}(g)

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

// Context is the graceful background master context exported for use where this
// background context should be extended to other processes or context wrappers
func (g *graceful) Context() context.Context { return g.ctx }

// Cancels the graceful background context and waits for a clean exit;
// is order flowed to abort multiple calls
func (g *graceful) Cancel() {
	if g.stop.CompareAndSwap(false, true) {
		g.framer("shutdown initiated")
		g.cancel() // signal manager shutdowns
		g.Shutdown()
	}
}

// Wait blocks until all .Init process have reported finished; ready state
func (g *graceful) Wait() {
	// delay timer to allow g.Init to register
	// at least one init.Add(1) event
	//time.Sleep(time.Millisecond * 250)
	g.init.Wait()
	g.framer("initilization complete")
}

// Register adds func() that are outside the .Init management
// architecture and that process before exiting via .Shutdown
func (g *graceful) Register(a ...func()) { g.register = append(g.register, a...) }

// Shutdown is order flow controlled to abort multiple calls and blocks on the background context
// and waits for all managed inits to terminate to cleanly exit; when a g.exit value is non-zero
// the process will call os.Exit(n), otherwise it just exits via a simple return; any additional
// registered func() will execute for controlled shutdown tasks outside the graceful architecture
func (g *graceful) Shutdown() {
	if g.wait.CompareAndSwap(false, true) { // ignore recurrent calls

		g.init.Wait()     // allow init bootstraps to complete
		<-g.ctx.Done()    // block and wait on context
		g.shutdown.Wait() // allow shutdowns to complete

		if g.bye.CompareAndSwap(false, true) { // ignore recurrent calls
			for i := range g.register {
				g.register[i]()
			}
			g.framer("bye")
			time.Sleep(time.Millisecond * 250)
			if g.exit != 0 {
				os.Exit(g.exit)
			}
		}
	}
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

	g.init.Add(len(obj) + 1)
	defer g.init.Done()

	for i := range obj {

		g.shutdown.Add(1)
		go func(obj interface{}, init *sync.WaitGroup) {
			defer g.shutdown.Done()
			switch fxn := obj.(type) {
			// func() expected to be non-blocking and init.Done()
			// tiggers afer the function returns; call to grace.Wait()
			// will confirm ready state
			case func():
				fxn()
				init.Done()
			// func(context.Context, *sync.WaitGroup) expected to block
			// and init.Done() triggers before context blocking occurs;
			// call to grace.Wait() confirms ready state
			case func(context.Context, *sync.WaitGroup):
				fxn(g.Context(), init)
			// func(context.Context) blocks on context, but can only signal
			// the process has started; a call to grace.Wait() will not confirm the
			// ready state, it can only signal Init started the process
			case func(context.Context):
				init.Done()
				fxn(g.Context())
			}

		}(obj[i], g.init)
		time.Sleep(time.Millisecond) // go routine ordering control

	}

	return g
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
