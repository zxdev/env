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

type Ab struct {
	A string `env:"a" default:"aa" help:"A is string"`
	B bool   `env:"b" help:"B is bool"`
	C int    `env:"c" help:"C is int"`
}

type Xy struct {
	X string `env:"x,require,environ,flagless" default:"xx" help:"X is string"`
	Y bool   `env:"y" help:"Y is bool"`
	Z int    `env:"z" help:"Z is int"`
}

/*

	graceful struct examples

*/

func (a *Ab) Start() env.GracefulFunc {
	time.Sleep(time.Second * 3)
	var count int
	log.Println("ab: running")
	return func(ctx context.Context) {
		for {
			time.Sleep(time.Second)
			select {
			case <-ctx.Done():
				return
			default:
				count++
				log.Println("sleep:", count)
			}
		}
	}
}

func (a *Ab) Starter() env.GracefulFunc {
	time.Sleep(time.Second * 3)
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

func (a *Ab) Function() func(ctx context.Context) {
	time.Sleep(time.Second * 2)
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

func (a *Ab) Param(param string) env.GracefulFunc {
	go func() {
		// note: with params, they must be wrapped or they do not advance
		// until their function completes and returns the GracefulFunc
		time.Sleep(time.Second)
	}()
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

func (a *Ab) Params(param string) func(ctx context.Context) {
	go func() {
		// note: with params, they must be wrapped or they do not advance
		// until their function completes and returns the GracefulFunc
		time.Sleep(time.Second * 3)
	}()
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

/*

	plain func examples

*/

func sample() env.GracefulFunc {
	time.Sleep(time.Second * 1)
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

func sample1(ctx context.Context) {
	time.Sleep(time.Second * 2)
	<-ctx.Done()
}

func sample2(param ...int) env.GracefulFunc {
	go func() {
		// note: with params, they must be wrapped or they do not advance
		// until their function completes and returns the GracefulFunc
		time.Sleep(time.Second)
	}()
	return func(ctx context.Context) {
		<-ctx.Done()
	}
}

/*

	test Shutdown; service loop
	test ReadyStop; once-through program

*/

func TestShutdown(t *testing.T) {

	var cfg1 Ab
	var cfg2 Xy

	env.Env()
	env.Init(&cfg1, &cfg2)
	defer env.Shutdown()

	env.Summary(&cfg1, &cfg2)
	env.Manage(&cfg1) // cfg.Start() GracefulFunc

	env.Ready()

}
func TestReadyStop(t *testing.T) {

	var cfg Ab
	cfg.A = "A"

	env.Init(&cfg)
	env.Summary(&cfg)

	log.Println("test: start/run all inits")
	env.Manage(&cfg)                           // cfg.Start() GracefulFunc
	env.Manage(cfg.Starter, "starter")         // cfg func GracefulFunc
	env.Manage(cfg.Param("param"), "param")    // cfg func(param) GracefulFunc
	env.Manage(cfg.Params("params"), "params") // cfg func(param) func(ctx context.Context)

	env.Manage(sample, "sample")       // func GracefulFunc
	env.Manage(sample1, "sample1")     // GracefulFunc
	env.Manage(sample2(42), "sample2") // func(param) GracefulFunc

	env.Ready()
	for i := 0; i < 3; i++ {
		log.Println("test: ...")
	}

	env.Stop()
}

// func TestWaitCalculation(t *testing.T) {

// 	checkOn := time.Hour
// 	t.Log(time.Now().Format(time.RFC3339))
// 	t.Log(time.Now().Add(checkOn).Sub(time.Now().Add(checkOn / 2).Round(checkOn)))
// 	t.Log(time.Now().Add(checkOn / 2).Round(checkOn).Sub(time.Now()))

// }
