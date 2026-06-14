# env — Executive Summary

`env` is a small, zero-dependency Go package for bootstrapping command-line
programs and long-running services. It replaces the boilerplate most Go daemons
end up rewriting — flag parsing, configuration, graceful lifecycle, and common
filesystem plumbing — with a single struct-tag-driven API and no dependencies
beyond the standard library.

## The problem it solves

A typical Go service stitches together several libraries and hand-rolled helpers
to: parse flags, read environment variables, load a config file, print
`--help`/`--version`, coordinate startup and signal-driven shutdown, hold a
single-instance lock, and clean up scratch files. `env` consolidates all of this
behind plain structs and struct tags, so a program declares *what* it needs and
the package wires up *how* it is sourced.

## What it provides

| Capability | Summary |
|------------|---------|
| Configuration | Populate plain structs from `default:` tags, environment variables, and command-line flags, with a defined override precedence. |
| CLI surface | Automatic `help` / `version` handling and a go-tool style subcommand dispatcher (`env.Commands`). |
| Flag forms | `-name value`, `-name=value`, `-name:value`, naked `-flag` for bools, aliases, and ordered positionals. |
| Graceful lifecycle | Signal-aware startup/shutdown controller that drives a master context and blocks until managed processes report done. |
| Filesystem utilities | Directory creation, JSON config loading, single-instance file lock, TTL scratch-file expiry, and gob persistence. |
| Environment paths | OS-aware standard directories (`/etc`, `/srv`, `/var`, `/tmp` on linux; a self-contained `_dev/` tree elsewhere). |

## Key characteristics

- **Zero dependencies** — standard library only; nothing to audit or update downstream.
- **Struct-tag driven** — configuration is declarative and lives next to the fields it describes.
- **Type support** — `string`, `bool`, `int`/`int64`, `uint`/`uint64`, `float32`/`float64`, and derived types such as `time.Duration`.
- **Small surface** — a handful of entry points (`NewEnv`/`Configure`, `Commands`, `NewGraceful`, plus utilities), easy to learn and to vendor.
- **Requires Go 1.26+.**

## When to use it

A good fit for in-house CLIs, daemons, and batch tools that want consistent
flag/config/lifecycle behavior without pulling in a larger framework. Less
suited to programs that need rich nested configuration, slice/map flags parsed
directly, or POSIX `--long`/`-short` GNU-style flag conventions — `env` favors a
deliberately compact convention over breadth.

## Where to go next

- **[User Guide](user-guide.md)** — complete, task-oriented documentation of every feature.
- **[README](../readme.md)** — project overview, install, and API summary.
- **[example/main.go](../example/main.go)** — a runnable subcommand CLI with a graceful service.
