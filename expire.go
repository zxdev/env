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
	CheckOn time.Duration   // default hourly
	report  bool            // report what was deleted
	path    []string        // directory targets
	age     []time.Duration // default daily
}

// NewExpire configurator will apply default settings
// and configure path items when provided
func NewExpire(path ...string) *Expire {

	expire := new(Expire)
	for i := range path {
		expire.Add(path[i], 0)
	}

	return expire
}

// Report toggles expiratin reporting on/off and returns new current setting; default off
func (ex *Expire) Report() bool { ex.report = !ex.report; return ex.report }

// Add will register a directory path and file age timeframe
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

// Start expire service checks for expired files periodically at the beginning
// of each CheckOn period; graceful
func (ex *Expire) Start(ctx context.Context) {

	if ex.CheckOn == 0 {
		ex.CheckOn = time.Hour
	}

	go func() {
		timer := time.NewTimer(time.Now().Add(ex.CheckOn / 2).Round(ex.CheckOn).Sub(time.Now()))
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				ex.Expire()
				return
			}
		}
	}()

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
