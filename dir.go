package env

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir will create the directory tree when it does not exist and return
// a string representation of the full composite path; fs.FileMode is 0755.
//
// WARNING: the final element is treated as a FILE (and therefore NOT created
// as a directory) whenever it contains any of the characters '.', '_' or '-'.
// This is a heuristic, not a guarantee: a directory whose name legitimately
// contains those characters (eg. "log_files", "my-app", "v1.2") will be
// MISTAKEN for a file, so only its parent is created and the intended
// directory is silently skipped — a later write into it will then fail.
//
//	env.Dir("srv", "cache")      // creates srv/cache        (dir)
//	env.Dir("srv", "data.db")    // creates srv only, returns srv/data.db (file presumed)
//	env.Dir("srv", "log_files")  // PITFALL: creates srv only — "log_files" presumed a file
//
// When the last element is a directory name containing ._- , append a trailing
// path element (eg. env.Dir("srv", "log_files", "")) or create it explicitly.
func Dir(a ...string) string {

	if len(a) > 0 {
		if strings.ContainsAny(a[len(a)-1], "._-") {
			os.MkdirAll(filepath.Join(a[:len(a)-1]...), 0755)
		} else {
			os.MkdirAll(filepath.Join(a...), 0755)
		}
	}

	return filepath.Join(a...)
}
