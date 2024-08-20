package main

// import "github.com/zxdev/env"

// /*

// go run example/main.go help

//  development
// --------------------
//  version
//  build

//  action          a     [or  ] default:           action chain[@path pull|process|expire|export]
//  secret                [   *] default:           the shared secret
//  show                  [    ] default:on         show the processing values

// */

// type Action struct {
// 	Action string    `env:"a,order,require" help:"action chain[@path pull|process|expire|export]"`
// 	Secret string    `env:"hidden" help:"the shared secret"`
// 	Show   bool      `default:"on" help:"show the processing values"`
// 	Seg    []string  `env:"-"` // args segments
// 	Path   *env.Path `env:"-"`
// }

// func main() {

// 	var a Action
// 	a.Path = env.NewEnv(&a)
// }
