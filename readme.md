# env package

Designed to provide simple and useful tooling to popoulate simple structs and provide managagement and helpful services utilizing struct tags instead of package like flag or third-party solutions.


```golang
type params struct {
	Action    string `env:"A,require,order" help:"a name to use"`
	Secret    string `env:"hidden" help:"a secret"`
	Flag      bool   `default:"on"  help:"a flag setting"`
	Number    int    `default:"5" help:"a number"`
	timestamp int64  // not parsed or reported in Summary
}

func main() {
	var param params
	paths := env.NewEnv(&param)
}

```

Set struct params and populate by calling ```env.NewEnv(&param)``` to parse and populate the struct as shown.
* Any default value is overloaded by system environment that is in turn overloaded by any command line values. 

Supported types in env.Parser are limited to ```string```, ```bool```, and ```int```. 
* Bool understands and accepts: ```on```, ```yes```, ```ok```, ```true```, and ```1``` and their associated negative counter parts. 
* Everything you want can be derived from these three basic types, including arrays and maps that utilize your own encoding and decoding.
	* Array can be passed or set as ```one,two,three``` and split by on comman, simarly a map can be encode as ```k1:v1,k1:v2``` and decoded by splitting on comma and then each set split on the colon.

---

Struct tag element supported and descriptions.

* ```env```: alias,order,require,environ,hidden
	* alias support can be short form of the switch ```-A``` instead of ```-action```
	* order makes it switchless and populated based on os.Args location index
	* require will cause hard stop when not defaulted or provided
	* environ sets all struct elements in the system envronment
	* hidden redacts the struct value in the summary report 

* ```default```: string, bool, int values
* ```help```: description

Automatic ```-help``` support reports basic information, the struct field name, the alias is any, the env:tag in use, any default value and the help description.

```
 % go run example/main.go -help

 development
--------------------
 version 
 build   

 action         A     [or  ] default:           an action to do
 secret               [   *] default:           a secret
 flag                 [    ] default:on         a flag setting
 number               [    ] default:5          a number

```

A summary log reports the struct values and integrates with other env system. If more than one param is populated by env.NewENV(&param,&server), each will appear as seperate sets in the order provided in the log summary output.

```

% go run example/main.go run   
2024/07/10 21:19:54 |----------------------------------------|
2024/07/10 21:19:54 | MAIN ::::::::::::::::::::::: event log |
2024/07/10 21:19:54 |-----//o--------------------------------|
2024/07/10 21:19:54                                 version
2024/07/10 21:19:54                                 build
2024/07/10 21:19:54                             pid 65812
2024/07/10 21:19:54 |-----//o--------------------------------|
2024/07/10 21:19:54  action         | run
2024/07/10 21:19:54  secret         | <hidden>
2024/07/10 21:19:54  flag           | true
2024/07/10 21:19:54  number         | 5
2024/07/10 21:19:54 |----------------------------------------|
2024/07/10 21:19:54 _dev/srv
2024/07/10 21:19:54 sample: start
2024/07/10 21:19:55 main: bootstrap complete
2024/07/10 21:19:58 main: interrupt shutdown
2024/07/10 21:19:58 sample: stop
2024/07/10 21:19:58 main: shutdown initiated
2024/07/10 21:19:58 |----------------------------------------|
2024/07/10 21:19:58 main: bye
2024/07/10 21:19:58 |----------------------------------------|


```

* env.NewEnv - parse and populate a param struct
* env.Parser - parser used with NewEnv methods

* env.Dir - ensure a directory exists
* env.Expire - expiration file manager with graceful interface support
* env.Graceful - graceful interface startup/shutdown controller
* env.Lock - process file lock (simple in use detection)
* env.Persist - persist and resume with data on disk
* env.Shutdown - shutdown, not necessary with graceful controller

