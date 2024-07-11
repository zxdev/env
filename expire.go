package env

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

/*

	var expire env.Expire
	expire.Add(nil,"my/expire/silent").Silent()
	expire.Silent().Add(nil, "my/silent/everything")
	...
	graceful.Manager(&expire)

*/

// Expire struct
type Expire struct {
	CheckOn time.Duration // frequency of checks (default: hourly)
	item    []struct {    // directory targets
		Path string
		TTL  time.Duration
	}
	silent bool
}

// Silent flag toggle for env.Expire, writes logs on os.Stderr (default: on)
func (ex *Expire) Silent() *Expire { ex.silent = !ex.silent; return ex }

// Add will register a directory/path with customized age timeframe (default: 24hr expiration)
func (ex *Expire) Add(ttl *time.Duration, path ...string) *Expire {

	if ttl == nil || *ttl == 0 {
		ttl24hr := time.Hour
		ttl = &ttl24hr // default
	}

	for i := range path {
		if len(path[i]) > 0 {
			ex.item = append(ex.item, struct {
				Path string
				TTL  time.Duration
			}{path[i], *ttl})
			if !ex.silent {
				log.Printf("expire: add %s ttl[%s]", filepath.Base(path[i]), *ttl)
			}
		}
	}

	return ex
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

// Expire will run the registered expiration processes
func (ex *Expire) Expire() *Expire {

	now := time.Now().Truncate(time.Second)
	for i := range ex.item {
		content, _ := os.ReadDir(ex.item[i].Path)
		for j := range content {
			if content[j].Type().IsRegular() {
				info, _ := os.Stat(filepath.Join(ex.item[i].Path, content[j].Name()))
				if !info.IsDir() && info.ModTime().Add(ex.item[i].TTL).Before(now) {
					if !ex.silent {
						log.Println("expire:", info.Name())
					}
					os.Remove(filepath.Join(ex.item[i].Path, info.Name()))
				}
			}
		}
	}

	return ex
}
