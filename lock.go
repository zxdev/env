package env

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

/*

	// single process lock
	var lock env.Lock
	if lock.Exist(nil) {
		return
	}
	lock.Lock()
	defer lock.Unlock()

*/

// Lock directory; default /tmp
type Lock string

// Exist reports the {file}.lock state as a boolean and
// expires the lock when past the ttl; default 1hr
func (lock *Lock) Exist(ttl *time.Duration) bool {

	if ttl == nil || *ttl == 0 {
		ttl1hr := time.Hour
		ttl = &ttl1hr // default
	}

	var path = string(*lock)
	if len(path) == 0 {
		path = "/tmp"
	}

	path = filepath.Join(path, filepath.Base(os.Args[0])+".lock")
	*lock = Lock(path)

	if _, err := os.Stat(filepath.Dir(path)); errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(filepath.Dir(path), 0755)
		return false
	}

	info, err := os.Stat(path)
	if info != nil && info.ModTime().Before(time.Now().Add(-(*ttl))) {
		return !lock.Unlock()
	}

	return !errors.Is(err, fs.ErrNotExist)
}

// Lock creates a {file}.lock and writes the current pid
func (lock Lock) Lock() bool {

	f, err := os.Create(string(lock))
	if err == nil {
		fmt.Fprint(f, os.Getpid())
		f.Close()
	}

	return err == nil
}

// Unlock removes a {file}.lock
func (lock Lock) Unlock() bool { return os.Remove(string(lock)) == nil }
