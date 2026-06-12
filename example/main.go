// Command example demonstrates env.Commands subcommand dispatch and the
// go-tool style help it renders. Try:
//
//	go run ./example                  # menu listing every command
//	go run ./example help             # same menu
//	go run ./example help serve       # serve's flag table
//	go run ./example serve help       # same flag table
//	go run ./example version          # one-line version
//	go run ./example pull -since 48h  # dispatch to pull
//	go run ./example serve -addr :9090 -tls
//
// The serve command additionally shows the graceful lifecycle driving a
// service to clean shutdown on os.Interrupt/SIGTERM/SIGHUP.
package main

import (
	"context"
	"log"
	"sync"

	"github.com/zxdev/env"
)

// pullCfg configures the pull subcommand; the env:/default:/help: tags drive
// both parsing and the help table.
type pullCfg struct {
	Since string `env:"s" default:"24h" help:"lookback window"`
	Limit int    `default:"100" help:"max records"`
}

// serveCfg configures the serve subcommand.
type serveCfg struct {
	Addr string `env:"a" default:":8080" help:"listen address"`
	TLS  bool   `help:"enable TLS"`
}

func main() {

	// Version/Build/Description populate the help banner; these are normally
	// set at build time with -ldflags, shown inline here for the demo.
	env.Version = "1.2.0"
	env.Build = "2026-06-11"
	env.Description = "example is a tool that demonstrates env subcommand dispatch."

	var pull pullCfg
	var serve serveCfg

	env.Commands(nil,
		env.Command{
			Name: "pull",
			Help: "retrieve and stage records",
			Cfg:  &pull,
			Run: func(path *env.Path) {
				log.Printf("pull: since=%s limit=%d -> %s", pull.Since, pull.Limit, path.Var)
			},
		},
		env.Command{
			Name: "serve",
			Help: "run the long-lived service",
			Cfg:  &serve,
			Run:  func(path *env.Path) { runServe(&serve) },
		},
	)
}

// runServe shows the graceful controller managing a service to clean shutdown.
func runServe(cfg *serveCfg) {

	log.Printf("serve: listening on %s (tls=%v)", cfg.Addr, cfg.TLS)

	var a Action
	grace := env.NewGraceful().Init(a.Start)
	defer grace.Shutdown()
	grace.Register(func() { log.Println("serve: extra cleanup") })
	grace.Wait()
}

// Action is a managed process for the graceful controller; the
// func(context.Context, *sync.WaitGroup) signature reports ready before it
// blocks, so grace.Wait() confirms the service is up.
type Action struct{}

func (a *Action) Start(ctx context.Context, init *sync.WaitGroup) {
	log.Println("action: start")
	defer log.Println("action: stop")
	init.Done()
	<-ctx.Done()
}
