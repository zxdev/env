package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// These var should be set externally by the build command
var (
	Version, Build string
	Description    string
)

// directory paths
type paths struct {
	Etc, Srv, Var, Tmp string
}

type env bool

// NewEnv that sets up the basic envrionment paths and
// calls the Parser to process the struct tag fields and
// populates any interfaces that are provided
//
//	type params struct {
//
//		env:"alias,order,require,environ,hidden"
//		help:"description"
//		default:"value"
//
//	Action string `env:"require" default:"server" help:"action [server|client]"`
//	}
//
//	supports bool, string, int types
func NewEnv(cfg ...interface{}) *paths {
	var e env
	return e.Configure(cfg...)
}

// NewEnvSilent sets up a silent environment
func NewEnvSilent(cfg ...interface{}) *paths {
	var e env = true // silent
	return e.Configure(cfg...)
}

// Configure sets up the basic environment
func (e *env) Configure(cfg ...interface{}) *paths {

	var env paths
	var name string
	switch runtime.GOOS {
	case "linux": // production
		env.Etc = "/etc"
		env.Srv = "/srv"
		env.Var = "/var"
		env.Tmp = "/tmp"
		name = filepath.Base(os.Args[0])

	default: // development
		env.Etc = "_dev/etc"
		env.Srv = "_dev/srv"
		env.Var = "_dev/var"
		env.Tmp = "_dev/tmp"
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
			for i := range cfg {
				e.helpTag(cfg[i])
			}
			fmt.Println()
			os.Exit(0)
		}
	}

	if len(cfg) > 0 {
		var p Parser
		p.Do(cfg...)
	}

	if !*e {

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
				if !v.Field(i).CanSet() || len(tag) == 0 || tag == "-" {
					continue
				}
				if hidden, ok := v.Type().Field(i).Tag.Lookup("env"); ok {
					if strings.Contains(hidden, "hidden") {
						log.Printf(" %-15s| <hidden>", strings.ToLower(v.Type().Field(i).Name))
						continue
					}
				}
				log.Printf(" %-15s| %v", tag, v.Field(i))
			}
			log.Printf("|%s|", strings.Repeat("-", 40))
		}

	}

	return &env
}

// helpTag displays help tags when present with struct field
func (e *env) helpTag(cfg interface{}) {

	// defer func() {
	// 	if recover() != nil {
	// 		fmt.Fprintln(os.Stderr, "\nhelp: interface misconfigured")
	// 		os.Exit(0)
	// 	}
	// }()

	var tag string
	var ok bool

	v := reflect.Indirect(reflect.ValueOf(cfg))
	for i := 0; i < v.NumField(); i++ {

		// if v.Field(i).Type().Kind() == reflect.Struct {
		// 	e.helpTag(v.Field(i).Interface())
		// 	continue
		// }

		// name field
		tag, ok = v.Type().Field(i).Tag.Lookup("name")
		if !ok {
			tag = strings.ToLower(v.Type().Field(i).Name)
		}
		if !v.Field(i).CanSet() || len(tag) == 0 || tag == "-" {
			continue
		}

		fmt.Printf(" %-15s", tag)

		var env struct{ Order, Require, Environ, Hidden, Alias string }
		if tag, ok := v.Type().Field(i).Tag.Lookup("env"); ok {
			for _, v := range strings.Split(tag, ",") {
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
		fmt.Printf("%-5s [%-1s%-1s%-1s%-1s] ",
			env.Alias, env.Order, env.Require, env.Environ, env.Hidden)

		// default field
		tag, _ = v.Type().Field(i).Tag.Lookup("default")
		fmt.Printf("default:%-10s ", tag)

		// help field
		tag, _ = v.Type().Field(i).Tag.Lookup("help")
		fmt.Println(tag)

	}
}
