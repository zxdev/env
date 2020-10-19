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
	"os"
	"time"

	"github.com/zxdev/env"
)

// Sample struct
type Sample struct {
	File string `env:"file,require,order" default:"sample.dat" help:"filename to use"`
	X    int    `help:"x is int"`
	Y    int
	z    int
}

// Start is the Graceful interface initializer
func (s *Sample) Start() env.GracefulFunc {
	// init code here
	return func(ctx context.Context) {
		<-ctx.Done()
		// shutdown code here
	}
}

// Run sample forever; graceful
func (s *Sample) Run() env.GracefulFunc {
	var index int
	return func(ctx context.Context) {
		for {

			time.Sleep(time.Second)

			select {
			case <-ctx.Done():
				return
			default:
			}

			index++
			log.Printf("%d sleep", index)
		}
	}

}

func main() {

	var sample Sample

	env.Env()              // mirror all exportable params in sample to os environment
	env.Fork(nil, &sample) // run normally or as a daemon enabled process

	f, _ := os.Create("fork.log")
	defer f.Close()
	log.SetOutput(f)

	env.Summary(&sample)

	env.Manage(&sample)
	env.Manage(sample.Run, "service")

	env.Ready()
	env.Shutdown()

}
