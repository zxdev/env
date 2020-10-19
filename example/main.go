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

package main

import (
	"context"
	"log"
	"time"

	"github.com/zxdev/env"
)

// Example struct
type Example struct {
	File    string `env:"file,require,order,environ" default:"sample.dat" help:"filename to use"`
	Service bool   `env:"service,require" help:"service loop flag"`
	X       int    `help:"x is int"`
	Y       int
	z       int
}

// Start is a the Graceful interface initializer
func (ex *Example) Start() env.GracefulFunc {
	// init code here
	return func(ctx context.Context) {
		<-ctx.Done()
		// shutdown code here
	}
}

func main() {

	var example Example
	env.Description = "An example program, MIT license."
	env.Init(&example)
	go env.Shutdown()

	/*
		f, _ := os.Create("demo.log")
		defer f.Close()
		log.SetOutput(f)
	*/

	env.Summary(&example)
	env.Manage(&example)

	if example.Service {
		// service loop example
		env.Ready()
		for {
			log.Println("...")
			time.Sleep(time.Second)
		}
		return
	}

	// once through example
	env.Ready()
	log.Println("...")
	env.Stop()

}
