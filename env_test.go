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

package env_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/zxdev/env"
)

/*

	graceful struct examples

*/

type Alpha struct {
	A string `env:"a" default:"aa" help:"A is string"`
	B bool   `env:"b" help:"B is bool"`
	C int    `env:"c" help:"C is int"`
}

func (a *Alpha) Start() func(ctx context.Context) {
	log.Println("alpha: init")
	time.Sleep(time.Second)
	var count int

	return func(ctx context.Context) {
		log.Println("alpha: running")
		for {
			time.Sleep(time.Second)
			select {
			case <-ctx.Done():
				return
			default:
				count++
				log.Println("alpah: sleep", count)
			}
		}
	}
}

type Beta struct {
	A string `env:"a" default:"aa" help:"A is string"`
	B bool   `env:"b" help:"B is bool"`
	C int    `env:"c" help:"C is int"`
}

func (a *Beta) Name() string { return "BETA" }
func (a *Beta) Start() func(ctx context.Context) {
	log.Println("beta: init")
	time.Sleep(time.Second)
	var count int

	return func(ctx context.Context) {
		log.Println("beta: running")
		for {
			time.Sleep(time.Second)
			select {
			case <-ctx.Done():
				return
			default:
				count++
				log.Println("beta: sleep", count)
			}
		}
	}
}

type Gamma struct {
	A string `env:"a" default:"aa" help:"A is string"`
	B bool   `env:"b" help:"B is bool"`
	C int    `env:"c" help:"C is int"`
}

func (a *Gamma) Closure() func(ctx context.Context) {
	log.Println("closure: init")
	time.Sleep(time.Second * 2)

	return func(ctx context.Context) {
		<-ctx.Done()
		log.Println("closure: done")
	}
}

func (a *Gamma) Basic(ctx context.Context) {
	log.Println("basic: init")
	time.Sleep(time.Second * 2)

	<-ctx.Done()
	log.Println("basic: done")
}

func (a *Gamma) Param1(zzz time.Duration) func(ctx context.Context) {
	log.Println("Param1: init")
	go func() {
		// note: with params, they must be wrapped or they do not advance
		// until their function completes which makes them lock-step ordered
		time.Sleep(zzz)
	}()
	return func(ctx context.Context) {
		<-ctx.Done()
		log.Print("Param1: done")

	}
}

/*

	plain func examples

*/

func sample1() func(ctx context.Context) {
	time.Sleep(time.Second * 1)
	log.Println("sample: 1!")
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

/*

	test Shutdown; service loop
	test ReadyStop; once-through program

*/

func TestShutdown(t *testing.T) {

	var alpha Alpha
	env.Env()
	env.Init(&alpha)
	env.Summary(&alpha)
	defer env.Shutdown()

	log.Println("bootstrap: begin")
	time.Sleep(time.Second)
	log.Println("bootstrap: finish")

	env.Ready()
	log.Println("exit: press ^C when ready")

}

func TestGraceful(t *testing.T) {

	var alpha Alpha
	alpha.A = "A"

	var beta Beta
	var gamma Gamma

	env.Init(&alpha)
	env.Summary(&alpha)

	log.Println("bootstrap: begin")
	env.NewExpire("/tmp/excite") // use defaults
	env.Manager(gamma.Param1(time.Second * 3))
	env.Manager(&alpha)
	env.Manager(&beta)
	env.Manager(gamma.Basic)
	env.Manager(gamma.Closure)
	env.Manager(sample1)
	log.Println("bootstrap: finish")

	env.Ready()
	log.Println("exit: stop in 2 seconds")
	time.Sleep(time.Second * 3)

	env.Stop()
}
