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

// TestNakedBool exercises the naked-bool flag form: a "-flag" with no value
// sets the bool true (at eof, before another flag, or before a bare token it
// must not swallow), the alias works the same, and -flag:off / -flag=false
// still disable it.
func TestNakedBool(t *testing.T) {

	type cfg struct {
		TLS   bool   `env:"t" help:"enable TLS"`
		Since string `env:"s" help:"a value flag"`
	}

	cases := []struct {
		name string
		args []string
		tls  bool
	}{
		{"naked at eof", []string{"prog", "-tls"}, true},
		{"naked before bare token", []string{"prog", "-tls", "foo"}, true},
		{"alias naked", []string{"prog", "-t"}, true},
		{"explicit colon off", []string{"prog", "-tls:off"}, false},
		{"explicit equals false", []string{"prog", "-tls=false"}, false},
		{"absent keeps default", []string{"prog"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args
			var c cfg
			env.NewEnv(&env.Options{NoExit: true, Silent: true}, &c)
			if c.TLS != tc.tls {
				t.Fatalf("TLS = %v, want %v", c.TLS, tc.tls)
			}
		})
	}

	// a naked bool must not swallow the following value flag: -tls leaves
	// -since to parse normally rather than consuming it as tls's value
	os.Args = []string{"prog", "-tls", "-since", "48h"}
	var c cfg
	env.NewEnv(&env.Options{NoExit: true, Silent: true}, &c)
	if !c.TLS || c.Since != "48h" {
		t.Fatalf("TLS=%v Since=%q, want true and %q", c.TLS, c.Since, "48h")
	}
}

// TestNakedBoolDefaultOff guards the default:"on" + -flag:off path: a naked
// presence-only flag never resets a defaulted-true bool.
func TestNakedBoolDefaultOff(t *testing.T) {

	type cfg struct {
		Quiet bool `default:"on" help:"suppress output"`
	}

	os.Args = []string{"prog"} // no flag -> default holds
	var c cfg
	env.NewEnv(&env.Options{NoExit: true, Silent: true}, &c)
	if !c.Quiet {
		t.Fatalf("Quiet = %v, want true (default:on)", c.Quiet)
	}

	os.Args = []string{"prog", "-quiet:off"}
	c = cfg{}
	env.NewEnv(&env.Options{NoExit: true, Silent: true}, &c)
	if c.Quiet {
		t.Fatalf("Quiet = %v, want false (-quiet:off)", c.Quiet)
	}
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
