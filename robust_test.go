package env_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zxdev/env"
)

// TestExpireTTLFailsafe guards the mass-deletion footgun: a ttl argument of an
// unsupported type (time.Duration is the natural mistake), or a zero, must fall
// back to the 24h default rather than registering TTL=0 and deleting everything.
func TestExpireTTLFailsafe(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "recent.txt")
	if err := os.WriteFile(keep, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	// age the file a couple seconds so a TTL=0 bug would expire it immediately
	old := time.Now().Add(-2 * time.Second)
	if err := os.Chtimes(keep, old, old); err != nil {
		t.Fatal(err)
	}

	// time.Duration is unsupported by the type switch -> must default to 24h
	var ex env.Expire
	ex.Silent().Add(time.Hour, dir).Expire()

	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("recent file deleted: unsupported ttl type fell through to TTL=0 (%v)", err)
	}
}

// TestPersistLoadNoPanic guards the nil-deref: Load on a missing file with a ttl
// must report "gone" and not panic.
func TestPersistLoadNoPanic(t *testing.T) {
	p := env.Persist(filepath.Join(t.TempDir(), "missing"))
	ttl := time.Hour
	var dst map[string]time.Time
	if !p.Load(&dst, &ttl) { // missing file -> treated as gone -> true
		t.Fatal("Load of missing file should return true (nothing to resume)")
	}
}

// TestPersistRoundTrip confirms Save/Load still works after the encode-error fix.
func TestPersistRoundTrip(t *testing.T) {
	p := env.Persist(filepath.Join(t.TempDir(), "state"))
	src := map[string]int{"a": 1, "b": 2}
	if !p.Save(src) {
		t.Fatal("Save failed")
	}
	var dst map[string]int
	if !p.Load(&dst, nil) {
		t.Fatal("Load failed")
	}
	if dst["a"] != 1 || dst["b"] != 2 {
		t.Fatalf("round trip mismatch: %v", dst)
	}
}

// TestGraceInitBadSignature guards the deadlock: an unsupported Init func
// signature must still release its init slot so Wait() returns.
func TestGraceInitBadSignature(t *testing.T) {
	done := make(chan struct{})
	go func() {
		grace := env.NewGraceful().Silent()
		grace.Init(func() error { return nil }) // unsupported signature
		grace.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Init with unsupported signature deadlocked Wait()")
	}
}
