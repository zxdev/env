package env

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Conf populates a json object applying tag:default conf values
// that are overloaded by the file source when configured at the
// primary level; no recurrsion support
//
//	type Example struct {
//		Text   string `json:"text,omitempty"`
//		Number int    `json:"number,omitempty" default:"10"`
//		Show bool     `json:"show,omitempty" default:"on"`
//	}
//
// supports: string, int, bool
func Conf(cfg interface{}, path string) {

	// conf.json {"text":"hello","number":5}
	// var cfg Example
	// env.Conf(&cfg, "conf.json")
	// t.Log(cfg)
	// 	=== RUN   TestConf
	//     conf_test.go:19: {hello 5 true}
	// --- PASS: TestConf (0.00s)

	v := reflect.Indirect(reflect.ValueOf(cfg))
	if v.Type().Kind() == reflect.Struct {
		for j := 0; j < v.NumField(); j++ {
			if s, ok := v.Type().Field(j).Tag.Lookup("default"); ok {
				switch v.Field(j).Kind() {
				case reflect.String:
					v.Field(j).SetString(s)
				case reflect.Int:
					n, _ := strconv.ParseInt(s, 10, 0)
					v.Field(j).SetInt(n)
				case reflect.Bool:
					switch strings.ToLower(s) {
					// case "off", "no", "false", "0":
					// 	v.Field(j).SetBool(false)
					case "on", "yes", "ok", "true", "1":
						v.Field(j).SetBool(true)
					}
				}
			}
		}
	}

	// load json object configuration file
	if len(path) > 0 {
		f, err := os.Open(path)
		if err == nil {
			json.NewDecoder(f).Decode(&cfg)
			f.Close()
		}
	}

}
