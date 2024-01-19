package env

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"time"
)

/*

	var persist env.Persist = "example"
	var m persist.NewMap()
	var ttl = time.Hour*24
	persist.Load(&m, ttl)
	m.Add("now_key")
	if next := m.Next(ttl); next != nil {
		var key string
		var more bool
		for {
			if key, more = next(); !more {
				break
			}
			// do stuff here
		}
	}
	if len(m) > 0 {
		persist.Save(m)
	}

*/

// Persist type
type Persist string

// filename verifies location and extension
func (p *Persist) filename() string {

	if !strings.HasSuffix(string(*p), ".persist") {
		*p += Persist(".persist")
	}

	return string(*p)
}

// Load persist object from disk or remove when older than stated ttl;
// ignores auto expiration when ttl is nil or 0
func (p Persist) Load(persist interface{}, ttl *time.Duration) bool {

	if ttl != nil && *ttl > 0 {
		info, err := os.Stat(p.filename())
		if os.IsNotExist(err) || info.ModTime().Before(time.Now().Add(-(*ttl))) {
			os.Remove(string(p))
			return true
		}
	}

	f, err := os.Open(p.filename())
	if err == nil {
		err = gob.NewDecoder(f).Decode(persist)
		f.Close()
	}

	return err == nil && os.Remove(string(p)) == nil
}

// Save persist object to disk; accepts anything
func (p Persist) Save(persist interface{}) bool {

	f, err := os.Create(p.filename())
	if err == nil {
		gob.NewEncoder(f).Encode(persist)
		f.Close()
	}
	fmt.Println(err)

	return err == nil
}

// Map of items with ttl
type Map map[string]time.Time

// NewMap
func NewMap() *Map {
	m := make(Map)
	return &m
}

// Add entry
func (m *Map) Add(k string) {
	if len(k) > 0 {
		(*m)[k] = time.Now()
	}
}

// Next returns a function return the key; removes key when used
// or when older than age, when age is non-zero
func (m *Map) Next(age time.Duration) func() (key string, more bool) {

	if len(*m) == 0 {
		return nil
	}

	return func() (string, bool) {
		for k := range *m {
			if age > 0 && (*m)[k].Before(time.Now().Add(-age)) {
				delete(*m, k)
				continue
			}
			delete(*m, k)
			return k, true
		}
		return "", false
	}
}
