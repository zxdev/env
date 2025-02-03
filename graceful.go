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
	silent             bool
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

// GraceInit starts a func and expectes a non-blocking func
// that exits when the init processes completes
//
//	func (a *Action) Init00() {
//		defer fmt.Println("init: complete")
//	}
//
// g.wgInit.Done() only triggers after the func returns
// and init must complete before a shutdown request is honored
func GraceInit(g *graceful, obj ...func()) *graceful {

	if g == nil {
		g = NewGraceful()
	}

	if !g.silent {
		log.Printf("%s: grace init [%d]", g.name, len(obj))
	}

	g.wgInit.Add(len(obj) + 1)
	defer g.wgInit.Done()

	for i := range obj {
		g.wgShutdown.Add(1)
		go func(i int) {
			obj[i]()
			g.wgInit.Done()
			g.wgShutdown.Done()
		}(i)
	}

	return g
}

// GraceInitContext starts the func and expects the blocking func to report
// when init process has completed and then blocks on <-ctx.Done(); both the
// Context and WaitGroup are supplied to the function from graceful
//
//	func (a *Action) Init01(ctx context.Context, init *sync.WaitGroup) {
//		log.Println("action: init01 entry")
//		defer log.Println("action: init01 exit")
//		init.Done()
//		<-ctx.Done()
//	}
//
// g.wgShutdown.Done() only triggers after the func returns
// and gracefule.wgInit.Done() must trigger once the init is complete
// and before contxt blocking or grace.Done() will never proceed
func GraceInitContext(g *graceful, obj ...func(context.Context, *sync.WaitGroup)) *graceful {

	if g == nil {
		g = NewGraceful()
	}

	if !g.silent {
		log.Printf("%s: grace initialize [%d]", g.name, len(obj))
	}

	g.wgInit.Add(len(obj) + 1)
	defer g.wgInit.Done()

	for i := range obj {
		g.wgShutdown.Add(1)
		go func(i int) {
			obj[i](g.Context(), g.wgInit)
			g.wgShutdown.Done()
		}(i)
	}

	return g
}

// Silent flag toggle for env.Graceful, writes logs on os.Stderr (default: on)
func (g *graceful) Silent() *graceful { g.silent = !g.silent; return g }

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
		if !g.silent {
			log.Printf("%s: shutdown initiated", g.name)
		}
		g.cancel() // signal manager shutdowns
		g.Wait()
	}
}

// Done blocks until all g.wgInit bootstaps are complete
func (g *graceful) Done() {
	// delay timer to allow g.Manager to register
	// at least one wgInit.Add(1) event
	time.Sleep(time.Millisecond * 250)
	g.wgInit.Wait()
	if !g.silent {
		log.Printf("%s: bootstrap complete", g.name)
	}
}

// Wait blocks on the g context and waits for inits to terminate to cleanly exit
// when a g.exit value is non-zero the process will call os.Exit(n), otherwise
// exits with a simple return and is order flow controlled to abort multiple calls
func (g *graceful) Wait() {
	if g.wait.CompareAndSwap(false, true) { // ignore recurrent calls

		g.wgInit.Wait()     // allow bootstraps to complete
		<-g.ctx.Done()      // block and wait on context
		g.wgShutdown.Wait() // allow shutdowns to complete

		if g.bye.CompareAndSwap(false, true) { // ignore recurrent calls
			if !g.silent {
				log.Printf("|%s|", strings.Repeat("-", 40))
				log.Printf(" %s: bye", g.name)
				log.Printf("|%s|", strings.Repeat("-", 40))
			}
			time.Sleep(time.Millisecond * 250)
			if g.exit != 0 {
				os.Exit(g.exit)
			}
		}
	}
}
