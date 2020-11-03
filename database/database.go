/*
Package database is for a local storage layer for the application
*/
package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v2"
)

var (
	db *badger.DB // The active database connection
)

// Connect will make a new database connection and new folder/file(s) if needed
func Connect(folder string) (err error) {

	// Get the home dir
	var home string
	if home, err = os.UserHomeDir(); err != nil {
		return err
	}

	// Set the database file and connect (disable logging for now)
	opts := badger.DefaultOptions(filepath.Join(home, folder, "database")).WithLogger(nil)
	// opts.EventLogging = false
	db, err = badger.Open(opts)
	return
}

// Disconnect will close the db connection
func Disconnect() error {
	return db.Close()
}

// Set will store a new key/value pair (expiration optional)
func Set(key, value string, ttl time.Duration) error {
	if db == nil {
		return fmt.Errorf("database is not connected")
	}
	return db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), []byte(value))
		if ttl > 0 {
			entry = entry.WithTTL(ttl)
		}

		return txn.SetEntry(entry)
	})
}

// Get will retrieve a value from a key (if found)
func Get(key string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database is not connected")
	}
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		valCopy, err = item.ValueCopy(nil)
		return err
	})

	// Not found (don't return an error, as we want to use this as cache)
	if errors.Is(err, badger.ErrKeyNotFound) {
		err = nil
	}

	return string(valCopy), err
}

// Flush will empty the entire database
func Flush() error {
	return db.DropAll()
}

// GarbageCollection will clean up some garbage in the database (reduces space, etc)
func GarbageCollection() error {
	err := db.RunValueLogGC(0.5)
	if errors.Is(err, badger.ErrNoRewrite) {
		return nil
	}
	return err
}
