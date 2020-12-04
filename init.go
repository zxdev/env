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

package env

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// default configuration setting
var (
	Identity    = filepath.Base(os.Args[0])          // Identity of app, as configured here
	Version     string                               // Version information, set by a builder.sh
	Build       string                               // Build information, set by a builder.sh
	Description string                               // Brief description, license, copyright
	EtcPath     Dir                         = "/etc" // EtcPath base path
	SrvPath     Dir                         = "/srv" // SrvPath base path
	VarPath     Dir                         = "/var" // VarPath base path
	development bool                                 // developtment flag
	env         bool                                 // env write settings to os.Environ
)

// Development flag toggle; apply development setting in Init()
func Development() bool { development = !development; return development }

// Env flag toggle; mirror all struct env:TAG=value to os environment via Parser()
func Env() bool { env = !env; return env }

// Init processe populates cfg structs by applying cfg struct default tag values,
// then any conf file (/etc/{identity}/{identity}.conf) values, then environment
// settings, then command line os.Args values to fill supported struct type
// fields; pass nil to load args or conf automatically
//
// configuration toggles:
//
//  Development() will force development settings that are otherwise autodetected
//  by the presense of a Development folder in the user home directory
//
//  Env() will mirror all final struct env:TAG=value to the os environment
func Init(cfg ...interface{}) {

	// autodetect dev system by presense of a Development folder in user home directory
	user, err := os.UserHomeDir()
	if err != nil {
		development = true
	}

	_, err = os.Stat(filepath.Join(user, "Development"))
	if !os.IsNotExist(err) || development {
		Identity = "development"
		Version = Identity
		Build = Identity
		EtcPath = ".dev/etc"
		SrvPath = ".dev/srv"
		VarPath = ".dev/var"
		development = true
	}

	Info(cfg...)
	Parser(nil, nil, env, cfg...)

}

// Fork is an alternative Init that enables a program to run normally or like
// a daemon start|stop process; pidPath directory must exist and the user must
// have r/w file level permissions for proper operation; pass nil for default
func Fork(pidPath *Dir, cfg ...interface{}) {

	if len(os.Args) > 1 {

		if pidPath == nil {
			pidPath = new(Dir)
		}

		if len(*pidPath) > 0 {
			pidPath.Create()
		}

		pidFile := pidPath.Join(Identity + ".pid")

		switch os.Args[1] {
		case "start":

			if _, err := os.Stat(pidFile); os.IsExist(err) {
				fmt.Fprintln(os.Stderr, "Already running!")
				os.Exit(0)
			}

			f, err := os.Create(pidFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create %s!\n", pidFile)
				os.Exit(0)
			}

			cmd := exec.Command(os.Args[0], os.Args[2:]...)
			if err = cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to start %s\n", filepath.Base(os.Args[0]))
				f.Close()
				os.Remove(pidFile)
				os.Exit(0)
			}

			f.WriteString(strconv.Itoa(cmd.Process.Pid))
			f.Close()
			fmt.Println(cmd.Process.Pid)
			os.Exit(0)

		case "stop":
			if _, err := os.Stat(pidFile); err != nil {
				fmt.Fprintln(os.Stderr, "Not running!")
				os.Exit(0)
			}

			data, _ := ioutil.ReadFile(pidFile)
			pid, err := strconv.Atoi(string(data))
			if pid == 0 || err != nil {
				fmt.Fprintf(os.Stderr, "Unable to parse %s\n", pidFile)
				os.Exit(0)
			}

			if process, err := os.FindProcess(pid); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to locate %s +%d\n", filepath.Base(os.Args[0]), pid)
			} else {
				if process.Signal(os.Interrupt) != nil {
					fmt.Fprintf(os.Stderr, "Unable to stop %s +%d\n", filepath.Base(os.Args[0]), pid)
				}
			}

			os.Remove(pidFile)
			os.Exit(0)

		}
	}

	Init(cfg...)

}

// Info on version or help request processor
//	prog version|-version|--version
//	prog help|-help|--help
func Info(cfg ...interface{}) {

	// 	 development
	// 	-----------------------
	// 	 version development
	// 	 build   development

	if len(os.Args) > 1 {

		var n = 18
		if len(Identity) > n {
			n = len(Identity)
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
				Identity, strings.Repeat("-", n+2), Version, Build)
			os.Exit(0)

		case "help":

			fmt.Printf("\n %-s\n%s\n version %s\n build   %s\n\n",
				Identity, strings.Repeat("-", n+2), Version, Build)
			if len(Description) > 0 {
				fmt.Printf("%s\n\n", Description)
			}
			for i := range cfg {
				helpTag(cfg[i])
			}
			fmt.Println()
			os.Exit(0)
		}
	}

}

// helpTag displays help tags when present with struct field
func helpTag(cfg interface{}) {

	defer func() {
		if recover() != nil {
			fmt.Println("info: interface is misconfigured")
			os.Exit(1)
		}
	}()

	v := reflect.Indirect(reflect.ValueOf(cfg))
	for i := 0; i < v.NumField(); i++ {

		if v.Field(i).Type().Kind() == reflect.Struct {
			helpTag(v.Field(i).Interface())
			continue
		}

		tag, ok := v.Type().Field(i).Tag.Lookup("env")
		if !ok && v.Field(i).CanSet() {
			tag = strings.ToLower(v.Type().Field(i).Name)
		}
		val, def := v.Type().Field(i).Tag.Lookup("default")
		help, ok := v.Type().Field(i).Tag.Lookup("help")
		if !ok || tag == "-" || len(tag) == 0 {
			continue
		}

		tag, special, _ := tagParse(tag)
		if len(special) > 0 {
			if len(val) > 0 {
				val += "] ["
			}
			val += special
			def = true
		}
		if def {
			help = fmt.Sprintf("%s [%s]", help, val)
		}
		if !strings.Contains(special, "order") {
			tag = "-" + tag
		}

		fmt.Printf(" %-15s | %-6s | %s\n", tag, v.Type().Field(i).Type.String(), help)
		// if len(special) > 0 {
		// 	fmt.Printf(" %15s |  :: %s\n", "", special)
		// }

	}
}

var summary bool

// Summary of cfg settings; log
func Summary(cfg ...interface{}) {

	summary = true
	log.Printf("|%s|", strings.Repeat("-", 40))
	log.Printf("| %s %s event log |", strings.ToUpper(Identity), strings.Repeat(":", 27-len(Identity)))
	log.Printf("|-----//o%s|", strings.Repeat("-", 32))
	log.Printf("%s%s version", strings.Repeat(" ", 31-len(Version)), Version)
	log.Printf("%s%s build", strings.Repeat(" ", 31-len(Build)), Build)
	log.Printf("%spid %d", strings.Repeat(" ", 28), os.Getpid())

	messageBar("configuration")
	for i := 0; i < len(cfg); i++ {
		envTag(cfg[i], "")
	}

	messageBar("service")

}

// messageBar formater
func messageBar(s string) { log.Printf("|---- %s -%so//---------|", s, strings.Repeat("-", 21-len(s))) }

// evnTag processor
func envTag(cfg interface{}, depth string) {

	defer func() {
		if recover() != nil {
			fmt.Fprintln(os.Stderr, "summary: interface is misconfigured")
			os.Exit(1)
		}
	}()

	v := reflect.Indirect(reflect.ValueOf(cfg))
	n := v.NumField()
	for i := 0; i < n; i++ {

		if v.Field(i).Type().Kind() == reflect.Struct {
			tag, _ := v.Type().Field(i).Tag.Lookup("env")
			envTag(v.Field(i).Interface(), depth+tag+".")
			continue
		}

		tag, ok := v.Type().Field(i).Tag.Lookup("env")
		if !ok && v.Field(i).CanSet() {
			tag = strings.ToLower(v.Type().Field(i).Name)
		}
		if tag == "-" || len(tag) == 0 {
			continue
		}

		tag = strings.SplitN(tag, ",", 2)[0]
		log.Printf("  %-15s| %v", depth+strings.ToLower(v.Type().Field(i).Name), v.Field(i))

	}

}

// Args processes os.Args and builds a m map[string]string; support for single
// reference switches -a aa -b=bb -c:cc formats; pass nil to create new
func Args(m map[string]string) map[string]string {

	if m == nil {
		m = make(map[string]string)
	}

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

	return m
}

// Conf processes a basic ini style file to build m map[string]string
// from the file; supports single reference k=v, k:v or k v setting; ignores
// comments and empty values; pass nil to create new
func Conf(path string, m map[string]string) map[string]string {

	if m == nil {
		m = make(map[string]string)
	}

	f, err := os.Open(path)
	if err != nil {
		return m
	}
	f.Close()

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
				m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				break
			}
		}
	}

	return m
}

// tag:env flag modifier
const (
	fRequire uint32 = 1 << iota // ,require field to have a value
	fOrder                      // ,order is inferred by os.Args position
	fEnviron                    // ,environ mirror value to os.Environ
)

// tagParse returns tag, text modifiers, and a composite flag set
//	require - ensure field has been given a value, no default
//	order - always orderly ordered values at the start
//	environ - mirror specifically to environment
func tagParse(s string) (string, string, uint32) {

	var flag uint32
	env := strings.SplitN(s, ",", 2)
	if len(env) == 1 {
		return env[0], "", 0
	}

	if strings.Contains(env[1], "require") {
		flag |= fRequire
	}
	if strings.Contains(env[1], "order") {
		flag |= fOrder
	}
	if strings.Contains(env[1], "environ") {
		flag |= fEnviron
	}
	return env[0], env[1], flag
}

// Parser will apply cfg struct default tag values, then any conf
// file (/etc/{identity}/{identity}.conf) values, then environment
// settings, followed by command line args values to fill supported
// struct type fields; pass nil to load args or conf automatically;
// set env=true to write KEY=value to os.Environ table
//	tag: env - name to use for configuration setting
//	tag: default - set default value
//	tag: help - help description
// supports string, bool, int, int64, uint, uint64 struct types
func Parser(args, conf map[string]string, env bool, cfg ...interface{}) {

	if args == nil {
		args = Args(nil)
	}

	if conf == nil {
		conf = Conf(EtcPath.Join(Identity, fmt.Sprintf("%s.conf", Identity)), nil)
	}

	for i := range cfg {
		fieldParser(cfg[i], args, conf, 1, env)
	}

}

// fieldParser will populate cfg fields
func fieldParser(cfg interface{}, args, conf map[string]string, order int, env bool) {

	defer func() {
		if recover() != nil {
			fmt.Fprintln(os.Stderr, "parser: interface is misconfigured")
			os.Exit(1)
		}
	}()

	v := reflect.Indirect(reflect.ValueOf(cfg))
	n := v.NumField()
	for j := 0; j < n; j++ {

		if v.Field(j).Type().Kind() == reflect.Struct {
			fieldParser(v.Field(j).Interface(), args, conf, order, env)
			continue
		}

		tag, ok := v.Type().Field(j).Tag.Lookup("env")
		if !ok && v.Field(j).CanSet() {
			// use name when not explicitly defined
			tag = strings.ToLower(v.Type().Field(j).Name)
		}
		if tag == "-" || len(tag) == 0 {
			continue
		}

		var value string
		var status bool

		// default tag settings; when defined
		if val, ok := v.Type().Field(j).Tag.Lookup("default"); ok {
			value, status = setField(v.Field(j), val)
		}

		var flag uint32
		tag, _, flag = tagParse(tag)

		// order, use os.Args only; unflagged order value extraction
		if flag&fOrder == fOrder && len(os.Args) > order {

			if !strings.HasPrefix(os.Args[order], "-") {
				value, status = setField(v.Field(j), os.Args[order])
				order++
			}

		} else {

			// conf map[string]string settings; A aa B=bb c:true
			if val, ok := conf[tag]; ok {
				value, status = setField(v.Field(j), val)
			}

			// environment settings; key is always upper case
			if val, ok := os.LookupEnv(strings.ToUpper(tag)); ok {
				value, status = setField(v.Field(j), val)
			}

			// Args map[string]string settings; -A aa -B=bb -c:true
			if val, ok := args[tag]; ok {
				value, status = setField(v.Field(j), val)
			}

		}

		if flag&fRequire == fRequire && !status {
			fmt.Fprintf(os.Stderr, "%s: missing required (%s) parameter\n", Identity, tag)
			os.Exit(0)
		}

		// mirror env:tag from struct to os environment table
		if env || flag&fEnviron == fEnviron && status {
			os.Setenv(tag, value)
		}

	}

}

// setField supports the string, bool, int, int64, uint, uint64 types
func setField(v reflect.Value, s string) (string, bool) {

	var ok bool

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
		ok = len(s) > 0

	case reflect.Bool:
		switch strings.ToLower(s) {
		case "on", "yes", "true", "1":
			v.SetBool(true)
			ok = true
		case "off", "no", "false", "0":
			v.SetBool(false)
			ok = true
		}

	case reflect.Int, reflect.Int64:
		n, _ := strconv.ParseInt(s, 10, 0)
		v.SetInt(n)
		ok = len(s) > 0 // accept 0 as valid

	case reflect.Uint, reflect.Uint64:
		n, _ := strconv.ParseUint(s, 10, 0)
		v.SetUint(n)
		ok = len(s) > 0 // accept 0 as valid

		//default:
		// unsupported, no-op
	}

	if !ok {
		s = ""
	}

	return s, ok
}
