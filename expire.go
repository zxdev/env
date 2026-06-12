package env

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type expire struct {
	Path string
	TTL  time.Duration
}

// Expire periodically removes regular files older than their registered TTL;
// Start satisfies the graceful Init signature so it can be managed directly.
//
//	var expire env.Expire
//	expire.Add(nil, "srv/cache")     // nil -> 24h TTL
//	expire.Add(6, "srv/short")       // int -> n hours
//	expire.Add("1h30m", "srv/build") // string -> parsed duration
//	expire.Freq = time.Hour          // sweep frequency (default hourly)
//	grace.Init(expire.Start)         // run under the graceful controller
type Expire struct {
	Freq   time.Duration // frequency of checks (default: hourly)
	item   []expire
	silent bool
}

// // NewExpire configurator
// func NewExpire(ttl any, path ...string) *Expire {
// 	var exp Expire
// 	exp.Add(ttl, path...)
// 	exp.Expire()
// }

// Silent flag toggle for env.Expire, writes logs on os.Stderr (default: on)
func (ex *Expire) Silent() *Expire { ex.silent = !ex.silent; return ex }

// Add will register a directory/path with customized age timeframe
// and supports various ttl inputs
//
//	nil          default 24h
//	int          n * hour
//	string       "24h", "1h30m"
func (ex *Expire) Add(ttl any, path ...string) *Expire {

	var exp time.Duration
	switch d := ttl.(type) {
	case nil:
		exp = time.Hour * 24
	case int:
		exp = time.Hour * time.Duration(d)
	case string:
		exp, _ = time.ParseDuration(d)
	}

	// failsafe: an unsupported ttl type, a zero, or a negative value would
	// otherwise register TTL=0 and delete every file on the next sweep
	if exp <= 0 {
		exp = time.Hour * 24
	}

	for i := range path {
		if len(path[i]) > 0 {
			ex.item = append(ex.item, expire{path[i], exp})
			if !ex.silent {
				log.Printf("expire: add %s ttl[%s]", filepath.Base(path[i]), exp)
			}
		}
	}

	return ex
}

// Start expire service manger to check for expired files periodically
// based on expire.Freq setting (default: check hourly, expire after 24hr)
func (ex *Expire) Start(ctx context.Context, init *sync.WaitGroup) {

	if ex.Freq == 0 { // use failsafe
		ex.Freq = time.Hour
	}

	ex.Expire()
	init.Done()

	var once bool
	ticker := time.NewTicker(time.Until(time.Now().Add(ex.Freq).Truncate(ex.Freq)))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if !once {
				once = !once
				ticker = time.NewTicker(ex.Freq)
			}
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
				if info != nil && !info.IsDir() && info.ModTime().Add(ex.item[i].TTL).Before(now) {
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
