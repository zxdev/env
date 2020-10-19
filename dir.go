// MIT License

// Copyright (c) 2020 zxdev

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package env

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir type
type Dir string

// Join appends a... and returns an updated string; no directory tree creation
func (d Dir) Join(a ...string) string { return filepath.Join(append([]string{string(d)}, a...)...) }

// Create appends a... and return an updated string; create the directory tree when it
// does not exist and return a string representation of the full composite path. A file
// is presumend when the last element contains any of the following ._- characters.
//
// 	conf.VarPath.Create() -> /var/log
// 	conf.VarPath.Create("insite") -> /var/log/insite
// 	conf.VarPath.Create("insite.log") ->  /var/log/insite.log
// 	conf.VarPath.Create("insite","insite.log") -> /var/log/insite/iniste.log
//
func (d Dir) Create(a ...string) string {

	defer func() {
		_, err := os.Stat(a[0])
		if os.IsNotExist(err) {
			os.MkdirAll(a[0], 0755)
		}
	}()

	// dir only
	if len(a) == 0 {
		a = []string{string(d)}
		return a[0]
	}

	// has a file ending the path
	if strings.ContainsAny(a[len(a)-1], "._-") {
		a = append([]string{string(d)}, a...)
		a[0] = filepath.Join(a[:len(a)-1]...)
		return filepath.Join(a[0], a[len(a)-1])
	}

	// does not have a file ending the path
	a = append([]string{string(d)}, a...)
	a[0] = filepath.Join(a...)
	return a[0]

}
