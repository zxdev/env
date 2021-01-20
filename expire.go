package env

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Expire is an file expiration manager
type Expire struct {
	CheckOn time.Duration   // frequency of checks (default: hourly)
	report  bool            // report what was deleted (default: off)
	path    []string        // directory targets
	age     []time.Duration // max age allowed (default: daily)
}

// NewExpire is an Expire wrapper that will create and start a file expiration
// manger under env.Manager control using default values applied to the direcory
// paths that are passed as paramaters.
//
// Configure *Expire directly when custom settings are desired.
func NewExpire(path ...string) {

	expire := new(Expire)
	for i := range path {
		expire.Add(path[i], 0)
	}
	Manager(expire)

}

// Start expire service manger to check for expired files periodically
// based on expire.CheckOn setting (default: check hourly, expire after 24hr)
func (ex *Expire) Start(ctx context.Context) {

	if ex.CheckOn == 0 { // use failsafe
		ex.CheckOn = time.Hour
	}
	ex.Expire()

	timer := time.NewTicker(ex.CheckOn)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			ex.Expire()
		}
	}

}

// Report toggles expiration reporting on/off (default: off)
func (ex *Expire) Report() *Expire { ex.report = !ex.report; return ex }

// Add will register a directory path with customized age timeframe
func (ex *Expire) Add(path string, age time.Duration) *Expire {

	if len(path) > 0 {
		ex.path = append(ex.path, path)
		if age == 0 {
			age = time.Hour * 24
		}
		log.Printf("expire: add %s @%s", filepath.Base(path), age)

		ex.age = append(ex.age, age)
	}

	return ex
}

// Expire will run the registered expiration processes
func (ex *Expire) Expire() *Expire {

	now := time.Now().Truncate(time.Second)
	for i := range ex.path {
		info, _ := ioutil.ReadDir(ex.path[i])
		for j := range info {
			if info[j].ModTime().Add(ex.age[i]).Before(now) {
				os.Remove(filepath.Join(ex.path[i], info[j].Name()))
				if ex.report {
					log.Println("expire:", ex.path[i], filepath.Base(info[j].Name()))
				}
			}
		}
	}

	return ex
}
