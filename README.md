# ENV PACKAGE

This is a handy configuration package that populates a struct with a limited set of type values, and provides graceful management controls for controlled startup and shutdown sequences as well as an ability to run a program like a daemon start|stop enabled process. 

```env package``` recognizes the following stuct field tags:

* `env:"name"` - name to use for the field value configuration; when omitted the exported field name is used.
* `default:"value"` - sets a default value for the struct field.
* `help:"description"` - helpful description text.

The tag:env processor also recognizes these special tag modifiers, and have the following effects:

* ```require``` - will ensure that a value has been provided to poputate the struct field. Setting a default tag meets this requirement and is duplicative, except where a required string value can not be left empty.
* ```order``` - will read args sequentialy from os.Args to populate the struct field in the native struct order, however order args must always appear before flagged args and follow the struct sequence order. Only a configured default tag value or an os.Args value will be recognized as valid value source.
* ```environ``` - will mirror and render the specific struct field to the os environment. This is unnecessary if env.Env() is toggled on since all fields would be mirrored to the os environment.

## Sample use

From example folder.

```golang

// Example struct
type Example struct {
  File  string `env:"file,require,order,environ" default:"sample.dat" help:"filename to use"`
  Block bool   `env:"B,require" help:"blocking flag"`
  X     int    `help:"x is int"`
  Y     int    `help:"y is int"`
  z     int
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
  env.Init(&example)
  go env.Shutdown()

  env.Summary(&example)
  env.Manage(&example)

  // service loop example
  if example.Block {
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

```

Note: The env package parser only supports string, bool, int, int64, uint, uint64 struct field types. The tag parser understands boolean values stated as true|false, yes|no, on|off, or 1|0. If more complex types are needed the developer will need to parse and/or type cast struct values to get them locally.

## Order of operations

The order of the application of the sources follows a lowest-to-highest archtype pattern and will overload previous values when the source key is present in the current processing source, this allows defaults to be applied and overloaded by higher order sources:

* apply tag default:"value" from the struct; when exists
* apply key:value settings from /etc/{identity}/{identity}.conf; when exists
* apply current os.Environ matching KEY= settings; when exists
* apply os.Args command line switches; when present

When the env.Env() toggle is called and set to true, all env:tag key=value pairs will be rendered and mirrored back to the os environment following the struct field archtype. Using the sample Ab struct from above, the following tag:env KEY=value pairs would be mirrored to the environment as:

```log
A=string B=string(bool) C=string(int)
```

Regardless of the key-value source for the order of operations, the Parser recognizes ```key value```, ```key=value```, or ```key:value``` to be valid key-value formats, as ahown below:

```log

$ A=abc ./prog -a xyz -b:on -c=42

2020/10/08 10:13:11 |----------------------------------------|
2020/10/08 10:13:11 | DEVELOPMENT :::::::::::::::: event log |
2020/10/08 10:13:11 |-----//o--------------------------------|
2020/10/08 10:13:11                     development version
2020/10/08 10:13:11                     development build
2020/10/08 10:13:11                             pid 20628
2020/10/08 10:13:11 |---- configuration ---------o//---------|
2020/10/08 10:13:11   A              | xyz
2020/10/08 10:13:11   B              | true
2020/10/08 10:13:11   C              | 42
2020/10/08 10:13:11 |---- service ---------------o//---------|

os environment [ A=xyz B=string(true) C=string(42) ]

```

## Variables

* Identity - of the app, os.Args[0] by default
* Version - information, set by a builder.sh
* Build - information, set by builder.sh
* Description - program description, license info, etc
* EtcPath, SrvPath, VarPath are base paths of Dir type

## Types

* Dir - type provides Join,Create path methods
* GracefulFunc - type is func(ctx context.Context)
* Graceful - interface type for Start() GracefulFunc
* Expire - path based file expiration manager

## Funcs

* Context - returns the background env package context
* DevOS - specify the development OS runtime environment; darwin by default
* Developement - toggle developement flag on|off; autodetected based on DevOS value
* Env - toggle env flag on|off to write finalized env:tag struct fields to os.Environ table
* Init - process, parse, and populate structs with values
* Fork - alternate Init that also allows a program to run like a daemon start|stop enabled process
* Summary - reports all exportable struct field values as well as graceful startup/shutdown information

exposed only for a customized Initialization alernative

* Info - provides version and help information
* Args - reads os.Args; use with Parse
* Conf - read ini style file; use with Parse
* Parser - populates all structs passed following the order or operations; structs must be pointer

## Graceful and GracefulFunc Management

The purpose of writing graceful functions and interfaces is to assure that packages and services are completely initialized before allowing the program to being normal operation, and to cleanly shutdown the same packages and services before exiting to avoid data loss and any other clean up tasks. All gracefully managed processes are go rountine wrapped so they will all start at the same time.

* Manage - wraps a GracefulFunc or Graceful interface to cleanly control startup/shutdown sequences
* Ready - blocks until all startup process have completed, the proceeeds
* Stop - signal all gracefully managed items to shutdown
* Shutdown - blocks and waits for a termination signal; Stop() or os.Interrupt, os.Kill signal


```golang
// GracefulFunc controller type
type GracefulFunc func(ctx context.Context)

// Graceful controller interface type
type Graceful interface { Start() GracefulFunc }

```

A gracefully managed item that needs params passed just needs to return a env.GracefulFunc, but it will block further execution until the GracefulFunc is returned. When lock-step initialization is desired, blocking is the desired behavior, however With a long or slow start sequence these may be go func(){...}() wrapped internally to prevent blocking of other initializations, see example below:

```golang

func pokey(path string, n int) env.GracefulFunc {
 go func() {
  time.Sleep(time.Second * 5)
 }()
 return func(ctx context.Context) {
  <-ctx.Done()
 }
}

```

Sample graceful sequence and output:

```log

env.Manage(&ab)
env.Manage(test)
env.Wait()

2020/10/09 10:15:17 |---- service ---------------o//---------|
2020/10/09 10:15:17 ab: start
2020/10/09 10:15:17 test: start
2020/10/09 10:15:20 |---- log -------------------o//---------|
2020/10/09 10:15:20 test: blah...
2020/10/09 10:15:24 |---- interrupt -------------o//---------|
2020/10/09 10:15:24 ab: stop
2020/10/09 10:15:24 test: stop
2020/10/09 10:15:24 |---- bye -------------------o//---------|

```

## Fork

Using Fork instead of Init allows a program to operate normally as well as run like a start|stop daemon process. This fires before Init would otherwise run, so the pidPath must be defined explicitly. If pidPath is not defined or is nil, the pid file will we written into the current directory.

```
 var cfg Ab
 Fork(nil,&cfg)
```

Usage: ```./example``` is normal foreground operation, while ```./example start``` and ```./example stop``` are the daemonized background forms. Any additional command line paramaters are just appending following the ```start``` keyword, and the start keyword will be removed during the forking process.

```log
$ ./example start -B:on
4209
$ ./example stop
```

## Version and Help Information

The package provides ```version``` information as well as a ```help``` configuration information drawn from the env:help tag when it is configured. Struct fields without the tag:help will not be displayed.

```log

$ ./example help

 development
-----------------------
 version development
 build   development

 -file           | filename to use [sample.dat]
                 |  :: require,sequence,environ
 -B              | blocking flag
                 |  :: require
 -x              | x is int
```


## Expiration manager

The file expiration manager, the ```Expire``` type, demonstrates the use of a graceful struct design and provides a configurable timebase automated directory cleanup service or process.

```golang

expire := env.NewExpire("path1","path2")
env.Manage(expire)

```

### Considerations

* All structs passed to Init, Fork, Summary, and Parser *must* be passed by pointer reference or the package will panic and report an interface misconfiguration. 

* All log output is directed to stdout by default. To redirect output to a log file call env.Init(), configure log.SetOutput(...), then call env.Summary() and the such. See example/fork.go for an example.

* Use env.Shutdown() and env.Ready() for service based programs, and incorporate env.Stop() with once-and-done based programs. See example/main.go for an example of each use case and configuration.

