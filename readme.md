# env package

Designed to provide simple and useful tooling to popoulate simple structs and provide managagement and helpful services.

* env.NewEnv - parse and populate a param struct
* env.Parser - parser used with NewEnv methods

* env.Dir - ensure a directory exists
* env.Expire - expiration file manager with graceful interface support
* env.Fork - process fork manager
* env.Graceful - graceful interface startup/shutdown controller
* env.Lock - process file lock (in use detection)
* env.Persist - persist simple data to diks
* env.Shutdown - shutdown, not necessary with graceful controller

See the ```example/main.go```for a sample use case.

```golang

// NewEnv that sets up the basic envrionment paths and
// calls the Parser to process the struct tag fields and
// populates any interfaces that are provided
//
//	type psarams struct {
//
//		env:"alias,require,order,environ"
//		default:"value"
//		help:"description"
//
//	Action string `env:"require" default:"server" help:"action [server|client]"`
//	}
//
//	supports bool, string, int types

// server.Server struct
type Server struct {
	Host     string `env:"H,require" default:"localhost" help:"localhost or FQDN"`
	Mirror   bool   `default:"on" help:"http request policy [mirror|400]"`
	CertPath string `default:"/var/certs"`
	opt      *http.Server
}

var srv server.Server
environ := env.NewEnv(&srv)

```