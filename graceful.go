package env

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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

// graceful struct
type graceful struct {
	wgBootstrap, wgShutdown *sync.WaitGroup
	ctx                     context.Context
	cancel                  context.CancelFunc
	silent                  bool
	name                    string
}

// NewGraceful configurator returns *graceful and starts the shutdown controller to
// capture (os.Interrupt, syscall.SIGTERM, syscall.SIGHUP) signals and waits on
// the <-graceful.context.Done() for a signal and waits for the graceful.Manager
// controller wgShutdown to confirm all managed processes and completed tasks before
// the program terminates execution
func NewGraceful() *graceful {

	g := new(graceful)
	g.wgBootstrap = new(sync.WaitGroup)
	g.wgShutdown = new(sync.WaitGroup)
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.name = filepath.Base(os.Args[0])

	go func(g *graceful) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		for {
			select {
			case <-g.ctx.Done():
				return
			case <-sig:
				signal.Stop(sig)
				g.Stop()
			}
		}
	}(g)

	return g
}

// Silent flag toggle for env.Graceful, writes logs on os.Stderr (default: on)
func (g *graceful) Silent() *graceful { g.silent = !g.silent; return g }

// Context is the graceful.context exported from the graceful manager for
// external use with processes not under the graceful.Manager controller
// that still need signaling to exit without g.wgShutdown reporting confirmation
func (g *graceful) Context() context.Context { return g.ctx }

// Cancel calls the graceful.context cancel() function; this function can be pass
// for external use with processes not under teh graceful.Manager controller for
// processes that require global termination signaling
func (g *graceful) Cancel() { g.cancel() }

// Done blocks until all graceful.Manager bootstaps are complete
func (g *graceful) Done() {
	// delay timer to allow graceful.Manager to register
	// at least one wgBootstrap.Add(1) event
	time.Sleep(time.Millisecond * 250)
	g.wgBootstrap.Wait()
	if !g.silent {
		log.Printf("%s: bootstrap complete", g.name)
	}
}

// Wait blocks on the graceful context and waits for bootstaps to terminate to cleanly exit
func (g *graceful) Wait() { g.bye() }

// Wait cancels the graceful context and waits for bootstaps to terminate to cleanly exit
func (g *graceful) Stop() {
	if !g.silent {
		log.Printf("%s: shutdown initiated", g.name)
	}
	g.cancel()
	g.bye()
}

// bye blocks on graceful context and waits for bootstaps to terminate to cleanly exit
func (g *graceful) bye() {
	g.wgBootstrap.Wait() // allow bootstraps to complete
	<-g.ctx.Done()       // block and wait on context
	g.wgShutdown.Wait()  // allow shutdowns to complete
	if !g.silent {
		time.Sleep(time.Millisecond * 250) // slight pause
		log.Printf("|%s|", strings.Repeat("-", 40))
		log.Printf("%s: bye", g.name)
		log.Printf("|%s|", strings.Repeat("-", 40))
	}
	os.Exit(0)
}

// Manager graceful controller configurator; structs with Start methods
// of specific signature types are supported
//
//	Start(ctx context.Context)
//	Start(ctx context.Context) error
//	Start(ctx context.Context, *sync.WaitGroup)
func (g *graceful) Manager(obj ...interface{}) {

	g.wgBootstrap.Add(1)
	defer g.wgBootstrap.Done()

	for i := range obj {

		g.wgBootstrap.Add(1)
		g.wgShutdown.Add(1)

		if reflect.TypeOf(obj[i]).Kind() != reflect.Ptr ||
			reflect.TypeOf(obj[i]).Elem().Kind() != reflect.Struct {
			fmt.Fprintf(os.Stderr, "%s: unsupported type", g.name)
			os.Exit(0)
		}

		name := strings.ToLower(reflect.TypeOf(obj[i]).Elem().Name())

		// object struct bootstrap signatures supported
		//  Start(ctx context.Context) error
		//  Start(ctx context.Context)
		//  Start(ctx context.Context, *sync.WaitGroup)

		switch object := obj[i].(type) {

		case interface {
			Start(context.Context)
		}: // Start(ctx context.Context)
			// expects a simple bootstrap, if any, that simply needs to
			// enter and remain in a loop or blocking on <-ctx.Done()
			// with or without any shutdown process task sequences
			go func() {
				if !g.silent {
					log.Printf("%s: start", name)
					defer log.Printf("%s: stop", name)
				}
				g.wgBootstrap.Done()
				object.Start(g.ctx)
				g.wgShutdown.Done()
			}()

		case interface {
			Start(context.Context) error
		}: // Start(ctx context.Context) error
			// expects the bootstrap process to complete and return
			// signaling the bootstrap has completed; hard exit on
			// any bootstrap failure
			go func() {
				if !g.silent {
					log.Printf("%s: start", name)
				}
				if err := object.Start(g.ctx); err != nil {
					log.Printf("%s: %s", name, err)
					os.Exit(0)
				}
				g.wgBootstrap.Done()
				g.wgShutdown.Done()
			}()

		case interface {
			Start(context.Context, *sync.WaitGroup)
		}: // Start(ctx context.Context, *sync.WaitGroup)
			// expects a bootstrap process to signal when complete and
			// then remain in a loop or blocking on <-ctx.Done() with
			// or without any shutdown process task sequences
			go func() {
				if !g.silent {
					log.Printf("%s: start", name)
					defer log.Printf("%s: stop", name)
				}
				object.Start(g.ctx, g.wgBootstrap)
				g.wgShutdown.Done()
			}()

		default:
			fmt.Fprintf(os.Stderr, "%s: unsupported struct", g.name)
			os.Exit(0) // hard stop
		}

	}
}
