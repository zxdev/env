package env

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Lock {file}.lock detection
type Lock struct {
	Path string        // lock directory
	TTL  time.Duration // default 1hr
}

// Unlock removes the current {file}.lock
func (lk *Lock) Unlock() bool {
	return os.Remove(filepath.Join(lk.Path, filepath.Base(os.Args[0])+".lock")) == nil
}

// Lock tests for the presence of a current {file}.Lock and returns true when
// a new {file}.Lock was established; false when an existing one is present
//
//	var lock = env.Lock{Path: "/tmp", TTL: time.Hour}
//	if !lock.Lock() {
//		return // existing lock
//	}
//	defer lk.Unlock()
func (lk *Lock) Lock() bool {

	// default assurances
	if lk.TTL == 0 {
		lk.TTL = time.Hour
	}
	if len(lk.Path) == 0 {
		lk.Path = "/tmp"
	}
	os.MkdirAll(filepath.Dir(lk.Path), 0755)

	// check existence and/or expired {file}.lock
	var target = filepath.Join(lk.Path, filepath.Base(os.Args[0])+".lock")
	info, _ := os.Stat(target) // verify exists
	if info != nil && info.ModTime().After(time.Now().Add(-lk.TTL)) {
		return false
	}

	// create {file}.lock
	f, err := os.Create(target)
	if err == nil {
		fmt.Fprint(f, os.Getpid())
		f.Close()
	}

	return err == nil
}
