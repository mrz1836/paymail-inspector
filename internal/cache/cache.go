// Package cache is the CLI's small, file-backed response cache.
//
// Paymail lookups (capabilities, PKI, SRV records, public profiles) are cached
// for a short time so repeated commands stay fast and polite to providers. The
// whole cache is one JSON document: it is loaded when the cache is opened,
// mutated in memory, and written back atomically on Close. That keeps it free
// of background goroutines and directory locks, so two terminals can run the
// CLI at the same time; when they do, the last one to exit wins, which is the
// right trade for a cache.
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// File and directory permissions for the cache. The cache only ever holds
// public paymail data, but there is no reason to share it with other users.
const (
	dirPerm  fs.FileMode = 0o700
	filePerm fs.FileMode = 0o600
)

// ErrInvalidKey is returned when a cache key is empty.
var ErrInvalidKey = errors.New("cache key must not be empty")

// Store is the cache contract the CLI depends on.
type Store interface {
	// Get decodes the live value stored under key into dst and reports
	// whether one was found. Expired and missing entries are both a miss.
	Get(key string, dst any) (bool, error)
	// Set stores value under key for ttl. A ttl of zero or less never expires.
	Set(key string, value any, ttl time.Duration) error
	// Flush removes every entry.
	Flush() error
	// Close persists pending changes and releases the store.
	Close() error
}

// entry is one cached value with its expiry.
type entry struct {
	Value   json.RawMessage `json:"value"`
	Expires time.Time       `json:"expires,omitzero"`
}

// expired reports whether e is no longer live at now.
func (e entry) expired(now time.Time) bool {
	return !e.Expires.IsZero() && !now.Before(e.Expires)
}

// File is a Store persisted to a single JSON file.
type File struct {
	path    string
	now     func() time.Time
	mu      sync.Mutex
	entries map[string]entry
	dirty   bool
}

// Option customizes a File.
type Option func(*File)

// WithClock replaces the time source, so tests can control expiry.
func WithClock(now func() time.Time) Option {
	return func(f *File) {
		f.now = now
	}
}

// Open loads the cache stored at path, creating its parent directory if
// needed. A missing file is an empty cache. A file that cannot be decoded is
// treated as empty too and is replaced on the next Close: losing a cache is
// never worth failing a command over.
func Open(path string, opts ...Option) (*File, error) {
	f := &File{
		path:    path,
		now:     time.Now,
		entries: map[string]entry{},
	}
	for _, opt := range opts {
		opt(f)
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return nil, fmt.Errorf("create cache directory: %w", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: the path is the CLI's own cache file, chosen by the CLI
	if errors.Is(err, fs.ErrNotExist) {
		return f, nil // a missing cache file is simply an empty cache
	}

	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var stored map[string]entry
	if json.Unmarshal(data, &stored) != nil {
		f.dirty = true // rewrite the unreadable file on Close

		return f, nil //nolint:nilerr // an unreadable cache is discarded, never fatal
	}

	if stored != nil {
		f.entries = stored
	}

	return f, nil
}

// Path returns the file backing the cache.
func (f *File) Path() string {
	return f.path
}

// Get implements Store.
func (f *File) Get(key string, dst any) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	e, ok := f.entries[key]
	if !ok {
		return false, nil
	}

	if e.expired(f.now()) {
		delete(f.entries, key)
		f.dirty = true

		return false, nil
	}

	if err := json.Unmarshal(e.Value, dst); err != nil {
		return false, fmt.Errorf("decode cached %q: %w", key, err)
	}

	return true, nil
}

// Set implements Store.
func (f *File) Set(key string, value any, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode cached %q: %w", key, err)
	}

	e := entry{Value: raw}
	if ttl > 0 {
		e.Expires = f.now().Add(ttl)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.entries[key] = e
	f.dirty = true

	return nil
}

// Len returns the number of live entries.
func (f *File) Len() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := f.now()
	count := 0

	for _, e := range f.entries {
		if !e.expired(now) {
			count++
		}
	}

	return count
}

// Flush implements Store. It empties the cache and deletes the backing file.
func (f *File) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.entries = map[string]entry{}
	f.dirty = false

	if err := os.Remove(f.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove cache: %w", err)
	}

	return nil
}

// Close implements Store. Expired entries are dropped and the cache is
// written atomically (temp file + rename) only if something changed.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.dirty {
		return nil
	}

	now := f.now()
	for key, e := range f.entries {
		if e.expired(now) {
			delete(f.entries, key)
		}
	}

	data, err := json.Marshal(f.entries)
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}

	if err = writeAtomic(f.path, data); err != nil {
		return err
	}

	f.dirty = false

	return nil
}

// writeAtomic replaces path with data so a reader never sees a partial file.
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp cache: %w", err)
	}

	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp cache: %w", err)
	}

	if err = tmp.Chmod(filePerm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp cache: %w", err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp cache: %w", err)
	}

	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}

	return nil
}

// Nop is a Store that remembers nothing. It stands in when the cache file
// cannot be opened, so commands keep working without caching.
type Nop struct{}

// Get implements Store; it always misses.
func (Nop) Get(string, any) (bool, error) { return false, nil }

// Set implements Store; it discards the value.
func (Nop) Set(string, any, time.Duration) error { return nil }

// Flush implements Store.
func (Nop) Flush() error { return nil }

// Close implements Store.
func (Nop) Close() error { return nil }
