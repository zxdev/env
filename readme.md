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

## Documentation

- **[Executive Summary](docs/executive-summary.md)** — what it is, the problems it solves, and when to use it.
- **[User Guide](docs/user-guide.md)** — complete, task-oriented documentation of every feature.
- **[example/main.go](example/main.go)** — a runnable subcommand CLI whose `serve` command drives a graceful service.

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
directories. Pass more than one struct to populate several at once. For
configuration precedence, struct tags, flag forms (including naked bools),
subcommands, the graceful lifecycle, and the utilities, see the
[User Guide](docs/user-guide.md).

---

## Highlights

- **Struct-tag configuration** — populate `string`, `bool`, `int`/`int64`,
  `uint`/`uint64`, `float32`/`float64`, and derived types (e.g. `time.Duration`)
  from `default:` tags, environment variables, and command-line flags, in a
  defined override order.
- **Flexible flags** — `-name value`, `-name=value`, `-name:value`, naked
  `-flag` for bools, aliases, and ordered positionals; automatic `help` /
  `version` handling.
- **Subcommands** — go-tool style dispatch via `env.Commands`.
- **Graceful lifecycle** — signal-aware startup/shutdown that drives a master
  context and blocks until managed processes report done.
- **Utilities** — directory creation, JSON config loading, single-instance file
  lock, TTL scratch-file expiry, and gob persistence.

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

See the [User Guide](docs/user-guide.md) for details on each.

---

## License

MIT — see [LICENSE](LICENSE).
