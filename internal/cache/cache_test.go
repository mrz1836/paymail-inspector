package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sample is a value shaped like the paymail responses the CLI caches.
type sample struct {
	Handle string `json:"handle"`
	PubKey string `json:"pubkey"`
}

// clock is a controllable time source.
type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func newClock() *clock {
	return &clock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
}

// openTemp opens a cache in a fresh temp directory.
func openTemp(t *testing.T, opts ...Option) (*File, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "nested", "cache.json")
	f, err := Open(path, opts...)
	require.NoError(t, err)

	return f, path
}

func TestOpen_MissingFileIsEmpty(t *testing.T) {
	f, path := openTemp(t)

	assert.Equal(t, path, f.Path())
	assert.Equal(t, 0, f.Len())

	var got sample

	found, err := f.Get("missing", &got)
	require.NoError(t, err)
	assert.False(t, found)

	// Nothing changed, so Close writes nothing.
	require.NoError(t, f.Close())
	assert.NoFileExists(t, path)
}

func TestFile_RoundTripAcrossOpen(t *testing.T) {
	f, path := openTemp(t)

	want := sample{Handle: "mrz@relayx.io", PubKey: "02abc"}
	require.NoError(t, f.Set("pki:mrz@relayx.io", want, time.Hour))
	assert.Equal(t, 1, f.Len())
	require.NoError(t, f.Close())

	info, err := os.Stat(path)
	require.NoError(t, err)

	if runtime.GOOS != "windows" {
		assert.Equal(t, filePerm, info.Mode().Perm())
	}

	reopened, err := Open(path)
	require.NoError(t, err)

	var got sample

	found, err := reopened.Get("pki:mrz@relayx.io", &got)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, want, got)
}

func TestFile_Expiry(t *testing.T) {
	c := newClock()
	f, path := openTemp(t, WithClock(c.Now))

	require.NoError(t, f.Set("short", sample{Handle: "a"}, time.Minute))
	require.NoError(t, f.Set("forever", sample{Handle: "b"}, 0))

	var got sample

	found, err := f.Get("short", &got)
	require.NoError(t, err)
	assert.True(t, found)

	c.now = c.now.Add(time.Minute) // exactly at expiry: no longer live

	assert.Equal(t, 1, f.Len())

	found, err = f.Get("short", &got)
	require.NoError(t, err)
	assert.False(t, found)

	found, err = f.Get("forever", &got)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "b", got.Handle)

	require.NoError(t, f.Close())

	data, err := os.ReadFile(path) //nolint:gosec // G304: test reads its own temp file
	require.NoError(t, err)
	assert.NotContains(t, string(data), "short")
	assert.Contains(t, string(data), "forever")
}

func TestFile_CloseDropsExpiredEntries(t *testing.T) {
	c := newClock()
	f, path := openTemp(t, WithClock(c.Now))

	require.NoError(t, f.Set("stale", sample{}, time.Second))

	c.now = c.now.Add(time.Hour)

	require.NoError(t, f.Close())

	reopened, err := Open(path)
	require.NoError(t, err)
	assert.Equal(t, 0, reopened.Len())
}

func TestFile_Flush(t *testing.T) {
	f, path := openTemp(t)

	require.NoError(t, f.Set("key", sample{}, time.Hour))
	require.NoError(t, f.Close())
	require.FileExists(t, path)

	require.NoError(t, f.Flush())
	assert.NoFileExists(t, path)
	assert.Equal(t, 0, f.Len())

	// Flushing an already empty cache is fine.
	require.NoError(t, f.Flush())
	require.NoError(t, f.Close())
	assert.NoFileExists(t, path)
}

func TestFile_InvalidKey(t *testing.T) {
	f, _ := openTemp(t)

	_, err := f.Get("", &sample{})
	require.ErrorIs(t, err, ErrInvalidKey)
	require.ErrorIs(t, f.Set("", sample{}, time.Hour), ErrInvalidKey)
}

func TestFile_DecodeErrors(t *testing.T) {
	f, _ := openTemp(t)

	require.NoError(t, f.Set("text", "a string", time.Hour))

	var wrongType sample

	found, err := f.Get("text", &wrongType)
	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), `decode cached "text"`)

	err = f.Set("func", func() {}, time.Hour)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `encode cached "func"`)
}

func TestOpen_CorruptFileIsReplaced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))

	f, err := Open(path)
	require.NoError(t, err)
	assert.Equal(t, 0, f.Len())
	require.NoError(t, f.Close())

	data, err := os.ReadFile(path) //nolint:gosec // G304: test reads its own temp file
	require.NoError(t, err)
	assert.JSONEq(t, "{}", string(data))
}

func TestOpen_NullFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	require.NoError(t, os.WriteFile(path, []byte("null"), 0o600))

	f, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, f.Set("key", sample{}, time.Hour))
	assert.Equal(t, 1, f.Len())
}

func TestOpen_Errors(t *testing.T) {
	t.Run("parent is a file", func(t *testing.T) {
		parent := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(parent, nil, 0o600))

		_, err := Open(filepath.Join(parent, "cache.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create cache directory")
	})

	t.Run("path is a directory", func(t *testing.T) {
		_, err := Open(t.TempDir())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read cache")
	})
}

func TestFile_CloseWriteError(t *testing.T) {
	f, path := openTemp(t)
	require.NoError(t, f.Set("key", sample{}, time.Hour))

	// Replace the cache directory with a file so the temp file cannot be
	// created next to the cache.
	dir := filepath.Dir(path)
	require.NoError(t, os.RemoveAll(dir))
	require.NoError(t, os.WriteFile(dir, nil, 0o600))

	err := f.Close()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create temp cache")
}

func TestFile_FlushError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions differ on windows")
	}

	f, path := openTemp(t)
	require.NoError(t, f.Set("key", sample{}, time.Hour))
	require.NoError(t, f.Close())

	dir := filepath.Dir(path)
	require.NoError(t, os.Chmod(dir, 0o500))       //nolint:gosec // G302: test makes its own temp dir read-only
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // G302: restore so the temp dir can be removed

	if os.Remove(filepath.Join(dir, "probe")) == nil {
		t.Skip("running with permissions that ignore directory modes")
	}

	err := f.Flush()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove cache")
}

func TestNop(t *testing.T) {
	var store Store = Nop{}

	found, err := store.Get("key", &sample{})
	require.NoError(t, err)
	assert.False(t, found)
	require.NoError(t, store.Set("key", sample{}, time.Hour))
	require.NoError(t, store.Flush())
	require.NoError(t, store.Close())
}

func BenchmarkFile_SetGet(b *testing.B) {
	f, err := Open(filepath.Join(b.TempDir(), "cache.json"))
	require.NoError(b, err)

	value := sample{Handle: "mrz@relayx.io", PubKey: "0352530c305378fd9dfd99f8c8c44e9092efa7c1674b61d4e9be65f92aa7a77bbe"}

	var got sample

	b.ReportAllocs()

	for b.Loop() {
		_ = f.Set("pki:mrz@relayx.io", value, time.Hour)
		_, _ = f.Get("pki:mrz@relayx.io", &got)
	}
}
