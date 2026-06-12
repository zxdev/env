package env_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/zxdev/env"
)

func TestEnv(t *testing.T) {

	type Action struct {
		Action string `env:"a,order,require" default:"pull" help:"action chain[@path pull|process|expire|export]"`
		Secret string `env:"hidden" help:"the shared secret"`
		Show   bool   `default:"on" help:"show the processing values"`

		Seg  []string  `env:"-"` // args segments
		Path *env.Path `env:"-"` // path params
	}

	var a Action
	a.Path = env.NewEnv(&a)

}

func TestHelp(t *testing.T) {

	type Action struct {
		Action string `env:"a,order,require" default:"pull" help:"action chain[@path pull|process|expire|export]"`
		Secret string `env:"hidden" help:"the shared secret"`
		Show   bool   `default:"on" help:"show the processing values"`

		Seg  []string  `env:"-"` // args segments
		Path *env.Path `env:"-"` // path params
	}

	// spoof help request
	os.Args = []string{"test", "help"}

	// we have to set opt.NoExit so this test will operate
	var a Action
	a.Path = env.NewEnv(&env.Options{NoExit: true}, &a)

}

func TestVersion(t *testing.T) {

	type Action struct {
		Action string `env:"a,order,require" default:"pull" help:"action chain[@path pull|process|expire|export]"`
		Secret string `env:"hidden" help:"the shared secret"`
		Show   bool   `default:"on" help:"show the processing values"`

		Seg  []string  `env:"-"` // args segments
		Path *env.Path `env:"-"` // path params
	}

	// spoof version request
	os.Args = []string{"test", "version"}
	env.Version = "test.0.0.0"
	env.Build = "abc"

	// we have to set opt.NoExit so this test will operate
	var a Action
	a.Path = env.NewEnv(&env.Options{NoExit: true}, &a)

}

type Action struct{}

func (a *Action) Start(ctx context.Context) {
	log.Println("action: start entry")
	defer log.Println("action: start exit")
	<-ctx.Done()
}

func (a *Action) Init00() {
	defer log.Println("action: init00")
}

func (a *Action) Init01(ctx context.Context, init *sync.WaitGroup) {
	log.Println("action: init01 entry")
	defer log.Println("action: init01 exit")
	init.Done()
	<-ctx.Done()
}

func (a *Action) Init02(ctx context.Context) {
	log.Println("action: init02 start")
	defer log.Println("action: init02 stop")
	time.Sleep(time.Second * 5)
	<-ctx.Done()
}

func TestGraceInit(t *testing.T) {

	var a Action
	grace := env.NewGraceful().Init(a.Init00, a.Init01)
	defer grace.Shutdown()
	grace.Register(func() { fmt.Println("bye-bye") })

	// generate a SIGTERM after 5s so the graceful controller exercises its
	// signal-driven shutdown path and the deferred Shutdown() unblocks rather
	// than waiting on a termination signal that never arrives
	go func() {
		time.Sleep(5 * time.Second)
		if p, err := os.FindProcess(os.Getpid()); err == nil {
			p.Signal(syscall.SIGTERM)
		}
	}()

	t.Log("grace.Done()")
	grace.Wait()

}
