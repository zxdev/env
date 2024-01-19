package env

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir will create the directory tree when it does not exist and return
// a string representation of the full composite path. A file is presumed
// when the last element contains any of the following ._- characters
// and fs.FileMode is coded to 0755
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
