package env_test

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/zxdev/env/v2"
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
	log.Println("action: init00 entry")
}

func (a *Action) Init01(ctx context.Context, init *sync.WaitGroup) {
	log.Println("action: init01 entry")
	defer log.Println("action: init01 exit")
	defer init.Done()
	<-ctx.Done()
}

func (a *Action) Init02(ctx context.Context) func() {
	log.Println("action: init02 entry")
	return func() {
		defer log.Println("action: init92 exit")
		<-ctx.Done()
	}
}

func (a *Action) Init03(ctx context.Context) func() {
	log.Println("action: init03 entry")
	return func() {
		go func() {
			defer log.Println("action: init03 exit")
			<-ctx.Done()
		}()
	}
}

func TestGraceInit(t *testing.T) {

	var a Action
	grace := env.GraceInit(nil, a.Init00)
	defer grace.Wait()

	grace.Done()

}

func TestGraceInitContext(t *testing.T) {

	var a Action
	grace := env.GraceInitContext(nil, a.Init01)
	defer grace.SetExit(-1).Wait()

	grace.Done()

}
