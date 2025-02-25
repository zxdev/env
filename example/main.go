package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/zxdev/env/v2"
)

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

func main() {

	var a Action
	grace := env.NewGraceful().Init(a.Init00, a.Init01, a.Init02)
	defer grace.Shutdown()
	grace.Register(func() { log.Println("extra: non-grace shutdown func") })
	grace.Wait()
}
