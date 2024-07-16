package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// These var should be set externally by the build command
var (
	Version, Build string
	Description    string
)

// NewEnv that sets up the basic envrionment paths and
// calls the Parser to process the struct tag fields and
// populates any interfaces that are provided
//
//	tags
//		env:"alias,order,require,environ,hidden"
//		help:"description"
//		default:"value" (bool, string, int)
//
//	Action string `env:"A,require" default:"server" help:"action [server|client]"`
func NewEnv(cfg ...interface{}) (path struct {
	Etc, Srv, Var, Tmp string
}) {
	return Configure(cfg...)
}

// Options for env.Configure
//
//	Silent: log configuration output
//	NoHelp: silences the help output
//	SetENV: set KEY=VALUE in environemnt
type Options struct {
	Silent bool // silence log configuration output
	NoHelp bool // silence help output
	SetENV bool // set KEY=VALUE in environment
}

// Configure sets up the basic environment and returns environment paths;
// pass Options as the first item to set or specify custom configuration
// options to silence log and help output and env.Options.M map populates,
// struct initially, overloaded by environment vars, overloaded by default
// tag, that is then overloaded by command line swithches, in this order
func Configure(cfg ...interface{}) (path struct {
	Etc, Srv, Var, Tmp string
}) {

	var opt Options
	switch c := cfg[0].(type) {
	case Options:
		opt = c
		cfg = cfg[1:]
	case *Options:
		opt = *c
		cfg = cfg[1:]
	}

	//var env paths
	var name string
	switch runtime.GOOS {
	case "linux": // production
		path.Etc = "/etc"
		path.Srv = "/srv"
		path.Var = "/var"
		path.Tmp = "/tmp"
		name = filepath.Base(os.Args[0])

	default: // development
		path.Etc = "_dev/etc"
		path.Srv = "_dev/srv"
		path.Var = "_dev/var"
		path.Tmp = "_dev/tmp"
		name = "development"
	}

	if len(os.Args) > 1 {

		var n = 18
		if len(name) > n {
			n = len(name)
		}
		if len(Version)+10 > n {
			n = len(Version) + 10
		}
		if len(Build)+10 > n {
			n = len(Build) + 10
		}

		switch strings.TrimLeft(os.Args[1], "-") {
		case "version":

			fmt.Printf("\n %-s\n%s\n version %s\n build   %s\n\n",
				name, strings.Repeat("-", n+2), Version, Build)
			os.Exit(0)

		case "help":

			fmt.Printf("\n %-s\n%s\n version %s\n build   %s\n\n",
				name, strings.Repeat("-", n+2), Version, Build)
			if len(Description) > 0 {
				fmt.Printf("%s\n\n", Description)
			}

			if !opt.NoHelp {
				for i := range cfg {

					var tag string
					var ok bool

					v := reflect.Indirect(reflect.ValueOf(cfg[i]))
					for j := 0; j < v.NumField(); j++ {

						// name field
						tag, ok = v.Type().Field(j).Tag.Lookup("name")
						if !ok {
							tag = strings.ToLower(v.Type().Field(j).Name)
						}
						if !v.Field(j).CanSet() || len(tag) == 0 {
							continue // unexported
						}

						var env struct{ Order, Require, Environ, Hidden, Alias string }
						if opts, ok := v.Type().Field(j).Tag.Lookup("env"); ok {
							if opts == "-" {
								continue
							}
							for _, v := range strings.Split(opts, ",") {

								switch v {
								case "order":
									env.Order = "o"
								case "require":
									env.Require = "r"
								case "environ":
									env.Environ = "e"
								case "hidden":
									env.Hidden = "*"
								default:
									env.Alias = v
								}
							}
						}
						// fmt.Printf(" %-15s", tag)
						fmt.Printf(" %-15s %-5s [%-1s%-1s%-1s%-1s] ",
							tag, env.Alias, env.Order, env.Require, env.Environ, env.Hidden)

						// default field
						tag, _ = v.Type().Field(i).Tag.Lookup("default")
						fmt.Printf("default:%-10s ", tag)

						// help field
						tag, _ = v.Type().Field(i).Tag.Lookup("help")
						fmt.Println(tag)

					}

				}
			}
			fmt.Println()
			os.Exit(0)
		}
	}

	if len(cfg) > 0 {
		opt.parse(cfg...)
	}

	if !opt.Silent {

		log.Printf("|%s|", strings.Repeat("-", 40))
		log.Printf("| %s %s event log |", strings.ToUpper(filepath.Base(os.Args[0])), strings.Repeat(":", 27-len(filepath.Base(os.Args[0]))))
		log.Printf("|-----//o%s|", strings.Repeat("-", 32))
		log.Printf("%s%s version", strings.Repeat(" ", 31-len(Version)), Version)
		log.Printf("%s%s build", strings.Repeat(" ", 31-len(Build)), Build)
		log.Printf("%spid %d", strings.Repeat(" ", 28), os.Getpid())
		log.Printf("|-----//o%s|", strings.Repeat("-", 32))

		var tag string
		var ok bool
		for j := 0; j < len(cfg); j++ {
			v := reflect.Indirect(reflect.ValueOf(cfg[j]))
			for i := 0; i < v.NumField(); i++ {
				if tag, ok = v.Type().Field(i).Tag.Lookup("name"); !ok {
					tag = strings.ToLower(v.Type().Field(i).Name)
				}
				if !v.Field(i).CanSet() || len(tag) == 0 {
					continue // unexported
				}
				if opts, ok := v.Type().Field(i).Tag.Lookup("env"); ok {
					if opts == "-" {
						continue
					}
					if strings.Contains(opts, "hidden") {
						log.Printf(" %-15s| <hidden>", strings.ToLower(v.Type().Field(i).Name))
						continue
					}
				}
				log.Printf(" %-15s| %v", tag, v.Field(i))
			}
			log.Printf("|%s|", strings.Repeat("-", 40))
		}

	}

	return
}

// parse will set the speficied cfg struct field value according to the tag:env and
// tag:default provided in the struct, and will overload in the following order:
//
//	tag:default, conf k:v sets, os.Args, os.Environ
//
// final values in the key:value os.Environment table.
//
//	env: alias,require,order,environ field flags
//	supports: string, bool, int/64, uint/64 types
func (p *Options) parse(cfg ...interface{}) {

	// overlaoding order
	// tag:default, conf, os.Args, ENV=

	var m = make(map[string]string)

	// processes os.Args and build/overload a map[string]string; support for single
	// reference switches -a aa -b
	for i := 0; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			key := strings.TrimLeft(os.Args[i], "-")
			switch {
			case strings.Contains(key, "="):
				s := strings.SplitN(key, "=", 2)
				m[s[0]] += s[1]
			case strings.Contains(key, ":"):
				s := strings.SplitN(key, ":", 2)
				m[s[0]] += s[1]
			default:
				i++
				if i < len(os.Args) {
					if !strings.HasPrefix(os.Args[i], "-") {
						m[key] = os.Args[i]
					} else {
						i--
					}
				}
			}
		}
	}

	// process interfaces
	for i := range cfg {

		var order = 1

		v := reflect.Indirect(reflect.ValueOf(cfg[i]))
		if v.Type().Kind() != reflect.Struct {
			fmt.Fprintf(os.Stderr, "%s: %s interface misconfigured",
				filepath.Base(os.Args[0]), reflect.TypeOf(cfg[i]).Elem().Name())
			os.Exit(1)
		}

		// process fields
		for j := 0; j < v.NumField(); j++ {

			// get field name
			name := strings.ToLower(v.Type().Field(j).Name)
			if !v.Field(j).CanSet() || len(name) == 0 {
				continue
			}

			var value string
			var status bool
			var env struct {
				Order, Require, Environ bool
				Alias                   string
			}

			// process tag:env
			if tag, ok := v.Type().Field(j).Tag.Lookup("env"); ok {
				if tag == "-" {
					continue // ignore
				}
				for _, v := range strings.Split(tag, ",") {
					switch v {
					case "order":
						env.Order = true
					case "require":
						env.Require = true
					case "environ":
						env.Environ = true
					// case "hidden":
					default:
						env.Alias = v
					}

				}
			}

			// apply tag:default values; when defined
			if val, ok := v.Type().Field(j).Tag.Lookup("default"); ok {
				value, status = p.setField(v.Field(j), val)
			}

			// overload with conf/args values; when present
			if val, ok := m[name]; ok {
				value, status = p.setField(v.Field(j), val)
			}
			if val, ok := m[env.Alias]; ok {
				value, status = p.setField(v.Field(j), val)
			}

			// overload with os.Environment table values; when present
			if val, ok := os.LookupEnv(strings.ToUpper(name)); ok {
				value, status = p.setField(v.Field(j), val)
			}

			// check for ordering
			if env.Order && len(os.Args) > order && !strings.HasPrefix(os.Args[order], "-") {
				// assumption is that we take args in order present to populate
				// the structure without using name flags {1} {2} {3} -blah
				value, status = p.setField(v.Field(j), os.Args[order])
				order++
			}

			// check for requiirement
			if env.Require && !status {
				fmt.Fprintf(os.Stderr, "%s: missing required (%s) parameter\n",
					filepath.Base(os.Args[0]), strings.ToLower(v.Type().Field(j).Name))
				os.Exit(0)
			}

			// mirror field NAME:VALUE from struct to the os.Environment table
			if status && (p.SetENV || env.Environ) {
				os.Setenv(name, value)
			}

		}

	}
}

// setField supports the string, bool, int, int64, uint, uint64 types as
// well as types derived from them (eg. time.Duration is int64); otherwise
// the field is ignored as nothing can be set
func (p *Options) setField(v reflect.Value, s string) (string, bool) {

	var ok bool

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
		ok = len(s) > 0

	case reflect.Int, reflect.Int64:
		n, _ := strconv.ParseInt(s, 10, 0)
		v.SetInt(n)
		ok = len(s) > 0 // accept 0 as valid

	case reflect.Uint, reflect.Uint64:
		n, _ := strconv.ParseUint(s, 10, 0)
		v.SetUint(n)
		ok = len(s) > 0 // accept 0 as valid

	case reflect.Bool:
		var value bool
		switch strings.ToLower(s) {
		//case "off", "no", "false", "0":
		case "on", "yes", "ok", "true", "1":
			value = true
			fallthrough
		default:
			v.SetBool(value)
			ok = true
		}

		//default:
		// unsupported, no-op
	}

	if !ok {
		s = ""
	}

	return s, ok
}
