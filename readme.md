# env

[![Go Reference](https://pkg.go.dev/badge/github.com/zxdev/env.svg)](https://pkg.go.dev/github.com/zxdev/env)

A small, zero-dependency Go package for bootstrapping command-line programs and
long-running services. It populates plain structs from struct tags, environment
variables, and command-line arguments, and bundles the runtime plumbing most
daemons end up rewriting: graceful startup/shutdown, environment paths, file
locks, on-disk persistence, and expiring scratch directories.

Everything is driven by struct tags instead of the `flag` package or a
third-party configuration library, and the whole package has no dependencies
beyond the standard library.

```bash
go get github.com/zxdev/env
```

Requires Go 1.26+.

---

## Quick start

```go
type params struct {
    Action string `env:"a,order,require" help:"action to perform"`
    Secret string `env:"hidden"          help:"a shared secret"`
    Flag   bool   `default:"on"          help:"a flag setting"`
    Number int    `default:"5"           help:"a number"`
    when   int64  // unexported: never parsed or reported
}

func main() {
    var p params
    path := env.NewEnv(&p) // parse, populate, and log a summary
    _ = path               // *env.Path with Etc/Srv/Var/Tmp
}
```

`env.NewEnv` (an alias for `env.Configure`) parses the tagged struct, prints a
summary log, and returns an `*env.Path` describing the environment's standard
directories. Pass more than one struct to populate several at once
(`env.NewEnv(&app, &server)`); each appears as its own block in the summary.

---

## Configuration

### Supported field types

`string`, `bool`, `int`, `int64`, `uint`, `uint64`, and any type derived from
them (for example `time.Duration`, which is an `int64`). Unsupported fields are
left untouched.

Booleans accept `on`, `yes`, `ok`, `true`, `1` as true; anything else is false.

Slices and maps are not parsed directly, but everything you need can be built
from the basic types: take a `string` and split it yourself
(`one,two,three` → slice; `k1:v1,k2:v2` → map).

### Overload precedence

For each field the value is resolved in this order, with later sources winning:

1. `default:` tag
2. command-line flag (`-name value`, `-name=value`, `-name:value`, or alias)
3. environment variable `NAME` (the field name, upper-cased)
4. ordered positional argument (when the field is tagged `order`)

### Struct tags

| Tag        | Purpose                                                                 |
|------------|-------------------------------------------------------------------------|
| `env`      | comma-separated flags: `alias`, `order`, `require`, `environ`, `hidden` |
| `default`  | initial value (`string`, `bool`, or `int`)                              |
| `help`     | description shown in `-help`                                             |
| `name`     | override the field label shown in the help table and summary log        |
| `env:"-"`  | ignore the field entirely (handy for embedded `*env.Path`, slices, etc.)|

`env` flag meanings:

- **alias** — a short switch, e.g. `-a` in place of `-action` (any token that
  isn't a reserved keyword is treated as the alias).
- **order** — switchless positional; populated by its index in `os.Args`.
- **require** — hard stop (exit code 1) when the field is neither defaulted nor
  provided.
- **environ** — mirror the resolved value back into the process environment
  (under the lower-cased field name).
- **hidden** — redact the value in the summary log (shown as `<hidden>`).

### Options

Pass an `*env.Options` (or `env.Options`) as the first argument to `Configure` /
`NewEnv` to control behavior:

```go
path := env.NewEnv(&env.Options{Silent: true, SetENV: true}, &p)
```

| Field    | Effect                                                              |
|----------|--------------------------------------------------------------------|
| `Silent` | suppress the summary log                                            |
| `NoHelp` | suppress the `-help` field table                                   |
| `SetENV` | mirror every resolved field into the environment (global `environ`)|
| `NoExit` | return `nil` instead of calling `os.Exit(0)` on version/help       |

`NoExit` is primarily for tests, so a spoofed `version`/`help` invocation
returns control instead of terminating the process.

### Build-time metadata

Set these package variables at build time to populate the banner and help
output:

```bash
go build -ldflags "\
  -X github.com/zxdev/env.Version=1.2.3 \
  -X github.com/zxdev/env.Build=$(git rev-parse --short HEAD)"
```

`env.Description` may also be set in code to add a paragraph to the help output.

### Environment paths

`Configure` detects the OS and returns standard directories:

| OS    | Etc        | Srv        | Var        | Tmp        |
|-------|------------|------------|------------|------------|
| linux | `/etc`     | `/srv`     | `/var`     | `/tmp`     |
| other | `_dev/etc` | `_dev/srv` | `_dev/var` | `_dev/tmp` |

On non-linux hosts the program runs in a self-contained `_dev/` tree, so
development never touches system directories.

### Runtime log control

The reserved `log` switch toggles timestamp headers at runtime without a
recompile: `-log on` enables date/time prefixes, `-log off` disables them. On
linux, timestamps are off by default (assuming an external logger adds them);
elsewhere they are on.

### Built-in commands

`version` and `help` are handled automatically. For the Quick-start `params`
struct, `help` prints a go-tool style `usage:` line and a compact flag table:

```text
% myprog help

usage: development [flags]

 action          a     [s] [or  ]            action to perform
 secret                [s] [   *]            a shared secret
 flag                  [b] [    ] (on)       a flag setting
 number                [i] [    ] (5)        a number
```

Each flag row is `name  alias  [type]  [ore*]  (default)  help`, where the type
code is one of `s` string, `i` int, `u` uint, `b` bool, `d` duration, and the
`[ore*]` slots mark `o`rder, `r`equire, `e`nviron, and hidden (`*`). A `default:`
tag is shown in parentheses. `version` prints a one-liner
(`name version <ver> build <build>`).

A populated run prints the summary log:

```text
% myprog run
2026/06/11 21:19:54 |----------------------------------------|
2026/06/11 21:19:54 | MYPROG ::::::::::::::::::::: event log |
2026/06/11 21:19:54 |-----//o--------------------------------|
2026/06/11 21:19:54                                 version
2026/06/11 21:19:54                                 build
2026/06/11 21:19:54                             pid 65812
2026/06/11 21:19:54 |-----//o--------------------------------|
2026/06/11 21:19:54  action         | run
2026/06/11 21:19:54  secret         | <hidden>
2026/06/11 21:19:54  flag           | true
2026/06/11 21:19:54  number         | 5
2026/06/11 21:19:54 |----------------------------------------|
```

---

## Subcommands

`env.Commands` adds go-tool style subcommand dispatch. Each `env.Command` pairs
a name and one-line help with a tagged config struct and an optional handler;
the selected command's flags are parsed with the same machinery as `Configure`.
This is the heart of [`example/main.go`](example/main.go):

```go
type pullCfg struct {
    Since string `env:"s" default:"24h" help:"lookback window"`
    Limit int    `default:"100" help:"max records"`
}

type serveCfg struct {
    Addr string `env:"a" default:":8080" help:"listen address"`
    TLS  bool   `help:"enable TLS"`
}

func main() {
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
            Run:  func(path *env.Path) { /* pull.exec(path) */ },
        },
        env.Command{
            Name: "serve",
            Help: "run the long-lived service",
            Cfg:  &serve,
            Run:  func(path *env.Path) { /* runServe(&serve) */ },
        },
    )
}
```

Dispatch surface:

```text
prog                 -> menu listing every command
prog help            -> menu
prog help <cmd>      -> that command's flag table
prog <cmd> help      -> that command's flag table
prog version         -> one-line version
prog <cmd> [flags]   -> Configure(cmd.Cfg) on the remaining args, then Run
```

The subcommand word is stripped from `os.Args` before parsing, so each command
sees only its own flags and ordered positionals. An unknown command prints the
menu and exits non-zero. Pass `*env.Options` (or `nil` for defaults) to control
`Silent`/`NoHelp`/`SetENV`/`NoExit`.

With no arguments (or `help`), `env.Commands` prints a go-tool style menu:

```text
% go run ./example

example is a tool that demonstrates env subcommand dispatch.

Usage:

	example <command> [arguments]

The commands are:

	pull   retrieve and stage records
	serve  run the long-lived service

Use "example help <command>" for more information about a command.
```

`help <cmd>` (or `<cmd> help`) drills into a command's flag table:

```text
% go run ./example help pull

usage: example pull [flags]

retrieve and stage records

 since           s     [s] [    ] (24h)      lookback window
 limit                 [i] [    ] (100)      max records

% go run ./example help serve

usage: example serve [flags]

run the long-lived service

 addr            a     [s] [    ] (:8080)    listen address
 tls                   [b] [    ]            enable TLS
```

`version` prints a one-liner, and any other first word dispatches to that
command — parsing its flags, logging the summary, then running its handler:

```text
% go run ./example version
development version 1.2.0 build 2026-06-11

% go run ./example pull -since 48h
2026/06/11 21:41:15 |----------------------------------------|
2026/06/11 21:41:15 | EXAMPLE :::::::::::::::::::: event log |
2026/06/11 21:41:15 |-----//o--------------------------------|
2026/06/11 21:41:15                           1.2.0 version
2026/06/11 21:41:15                      2026-06-11 build
2026/06/11 21:41:15  mac                        pid 20175
2026/06/11 21:41:15 |-----//o--------------------------------|
2026/06/11 21:41:15  since          | 48h
2026/06/11 21:41:15  limit          | 100
2026/06/11 21:41:15 |----------------------------------------|
2026/06/11 21:41:15 pull: since=48h limit=100 -> _dev/var
```

---

## Graceful lifecycle

`env.Graceful` coordinates startup and shutdown for long-running services. It
captures `os.Interrupt`, `SIGTERM`, and `SIGHUP`, drives a master context, and
blocks until every managed process has reported done.

```go
func main() {
    var a Action
    grace := env.NewGraceful().Init(a.Init00, a.Init01, a.Init02)
    defer grace.Shutdown()
    grace.Register(func() { log.Println("extra cleanup") })
    grace.Wait() // block until all Init processes report ready
}
```

`Init` accepts three function signatures:

| Signature                                | Behavior                                                                 |
|------------------------------------------|--------------------------------------------------------------------------|
| `func()`                                 | non-blocking; `init.Done()` fires when it returns. `Wait()` confirms ready. |
| `func(context.Context, *sync.WaitGroup)` | blocking; call `init.Done()` yourself when ready, then block on `<-ctx.Done()`. `Wait()` confirms ready. |
| `func(context.Context)`                  | blocking; ready state is indeterminate — `Wait()` only confirms it started. |

Control methods:

- `Silent()` — toggle the framed event logging (default on).
- `Frame()` — toggle the `|----|` bar frames around events.
- `SetExit(n)` — exit code for `os.Exit(n)`; `0` (default) exits via a plain return.
- `Context()` — the master `context.Context`, for extending to other processes.
- `Register(fns...)` — extra cleanup functions run during shutdown, outside the Init architecture.
- `Wait()` — block until all Init processes report ready.
- `Cancel()` — trigger an orderly shutdown programmatically.
- `Shutdown()` — block until context is done and all managed processes exit.

`env.Shutdown(ctx, fn)` is a minimal alternative for programs that don't use the
full graceful controller: it blocks on a signal or the context, runs `fn`, and
calls `os.Exit(0)`.

---

## Utilities

### `env.Dir` — ensure a directory tree

```go
env.Dir("srv", "cache")   // creates srv/cache, returns the joined path
env.Dir("srv", "data.db") // creates srv only; treats data.db as a file
```

> **Pitfall:** the final element is treated as a *file* (and not created as a
> directory) when it contains `.`, `_`, or `-`. A real directory named
> `log_files`, `my-app`, or `v1.2` is mistaken for a file and silently skipped.
> Append a trailing element (`env.Dir("srv", "log_files", "")`) or create it
> explicitly to work around the heuristic.

### `env.Conf` — JSON config with defaults

Apply `default:` tags, then overload from a JSON file (top level only, no
recursion):

```go
type Example struct {
    Text   string `json:"text,omitempty"`
    Number int    `json:"number,omitempty" default:"10"`
    Show   bool   `json:"show,omitempty"   default:"on"`
}

var cfg Example
env.Conf(&cfg, "conf.json") // missing/unreadable file leaves defaults intact
```

### `env.Lock` — single-instance file lock

```go
lock := env.Lock{Path: "/tmp", TTL: time.Hour}
if !lock.Lock() {
    return // another instance holds a fresh lock
}
defer lock.Unlock()
```

Writes `{program}.lock` containing the PID into `Path`. A lock older than `TTL`
(default 1h) is treated as stale and reacquired.

### `env.Expire` — expiring scratch directories

Periodically removes regular files older than their TTL. Implements the
graceful `Init` signature, so it can be managed directly.

```go
var expire env.Expire
expire.Add(nil, "srv/cache")     // nil → 24h TTL
expire.Add(6, "srv/short")       // int → n hours
expire.Add("1h30m", "srv/build") // string → parsed duration
expire.Freq = time.Hour          // sweep frequency (default hourly)

grace.Init(expire.Start)
```

An invalid, zero, or negative TTL falls back to 24h rather than deleting
everything on the next sweep. `Silent()` toggles logging.

### `env.Persist` — gob persistence with TTL

Save and resume arbitrary data across runs; an optional TTL expires stale state.

```go
var store env.Persist = "queue" // -> queue.persist
ttl := 24 * time.Hour

m := env.NewMap()
store.Load(m, &ttl) // older than ttl? removed instead of loaded
m.Add("job-key")

if next := m.Next(ttl); next != nil {
    for {
        key, more := next()
        if !more {
            break
        }
        // process key (consumed as it is read)
    }
}

if len(*m) > 0 {
    store.Save(m) // persist what is left
}
```

`env.Map` (`map[string]time.Time`) is a convenience type whose `Next` iterator
drains entries as it yields them and drops anything older than the supplied age.
`Load`/`Save` accept any gob-encodable value, not just `Map`.

---

## API summary

| Symbol                         | Purpose                                              |
|--------------------------------|------------------------------------------------------|
| `env.NewEnv` / `env.Configure` | parse & populate tagged structs, return `*env.Path`  |
| `env.Commands`                 | go-tool style subcommand dispatch                    |
| `env.Options`                  | configure Silent/NoHelp/SetENV/NoExit                |
| `env.Path`                     | standard environment directories (Etc/Srv/Var/Tmp)   |
| `env.Conf`                     | JSON config file with `default:` tag support         |
| `env.Dir`                      | ensure a directory tree exists                       |
| `env.Lock`                     | single-instance file lock                            |
| `env.Expire`                   | TTL-based scratch-file cleanup (graceful `Init`)     |
| `env.Persist` / `env.Map`      | gob persistence with optional expiry                 |
| `env.NewGraceful`              | graceful startup/shutdown controller                 |
| `env.Shutdown`                 | minimal signal/context shutdown helper               |

See [`example/main.go`](example/main.go) for a runnable subcommand CLI whose
`serve` command drives a graceful service.

---

## License

MIT — see [LICENSE](LICENSE).
