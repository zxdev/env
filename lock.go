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

// Lock sets a current {file}.lock or expires and exiting based on lock.TTL
// and returns true a lock was successfully set
//
//	var lk = lock.Lock{Path: "/tmp", TTL: time.Hour}
//	if lk.Lock() {
//		return
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
	info, err := os.Stat(target)
	if info != nil { // exists
		if !info.ModTime().Before(time.Now().Add(-lk.TTL)) {
			return true
		}
	}

	// create {file}.lock
	f, err := os.Create(target)
	if err == nil {
		fmt.Fprint(f, os.Getpid())
		f.Close()
	}

	return err != nil
}
