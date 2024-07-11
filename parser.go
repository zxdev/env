package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type Parser struct {
	ConfPath *[]string
	SetENV   bool
	m        map[string]string
}

// Do will set the speficied cfg struct field value according to the tag:env and
// tag:default provided in the struct, and will overload in the following order:
//
//	tag:default, conf k:v sets, os.Args, os.Environ
//
// final values in the key:value os.Environment table.
//
//	env: alias,require,order,environ field flags
//	supports: string, bool, int/64, uint/64 types
func (p *Parser) Do(cfg ...interface{}) {

	// overlaoding order
	// tag:default, conf, os.Args, ENV=

	if p.m == nil {
		p.m = make(map[string]string)
	}

	// processes a basic ini style file to build map[string]string
	// from the file; supports single reference k=v, k:v or k v setting; ignores
	// comments and empty values; pass nil etcPath to skip
	if p.ConfPath != nil && len(*p.ConfPath) > 0 {
		for i := range *p.ConfPath {
			f, err := os.Open(filepath.Join((*p.ConfPath)[i], filepath.Base(os.Args[0]), filepath.Base(os.Args[0])+".conf"))
			if err == nil {
				sep := []string{"=", ":", " "}
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					s := strings.TrimSpace(scanner.Text())
					if len(s) == 0 || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "//") {
						continue
					}
					for i := range sep {
						if strings.Contains(s, sep[i]) {
							kv := strings.SplitN(s, sep[i], 2)
							p.m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
							break
						}
					}
				}
				f.Close()
			}
		}
	}

	// processes os.Args and build/overload a map[string]string; support for single
	// reference switches -a aa -b
	for i := 0; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			key := strings.TrimLeft(os.Args[i], "-")
			switch {
			case strings.Contains(key, "="):
				s := strings.SplitN(key, "=", 2)
				p.m[s[0]] += s[1]
			case strings.Contains(key, ":"):
				s := strings.SplitN(key, ":", 2)
				p.m[s[0]] += s[1]
			default:
				i++
				if i < len(os.Args) {
					if !strings.HasPrefix(os.Args[i], "-") {
						p.m[key] = os.Args[i]
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
			if name == "-" || len(name) == 0 {
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
				for _, v := range strings.Split(tag, ",") {
					switch v {
					case "order":
						env.Order = true
					case "require":
						env.Require = true
					case "environ":
						env.Environ = true
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
			if val, ok := p.m[name]; ok {
				value, status = p.setField(v.Field(j), val)
			}
			if val, ok := p.m[env.Alias]; ok {
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
func (p *Parser) setField(v reflect.Value, s string) (string, bool) {

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
