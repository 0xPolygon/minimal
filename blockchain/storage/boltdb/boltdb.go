package boltdb

import (
	"fmt"
	"path/filepath"

	"github.com/0xPolygon/minimal/blockchain/storage"
	"github.com/boltdb/bolt"
	"github.com/hashicorp/go-hclog"
)

// Factory creates a boltdb storage
func Factory(config map[string]interface{}, logger hclog.Logger) (storage.Storage, error) {
	path, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("path not found")
	}
	pathStr, ok := path.(string)
	if !ok {
		return nil, fmt.Errorf("path is not a string")
	}
	return NewBoltDBStorage(filepath.Join(pathStr, "db"), logger)
}

// NewBoltDBStorage creates the new storage reference with boltdb
func NewBoltDBStorage(path string, logger hclog.Logger) (storage.Storage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{})
	if err != nil {
		return nil, err
	}

	kv := &boltDBKV{db}
	return storage.NewKeyValueStorage(logger, kv), nil
}

// boltDBKV is the boltdb implementation of the kv storage
type boltDBKV struct {
	db *bolt.DB
}

var bucket []byte = []byte{'b'}

func (l *boltDBKV) Set(p []byte, v []byte) error {
	err := l.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(p, v)
	})
	return err
}

func (l *boltDBKV) Get(p []byte) ([]byte, bool, error) {
	var data []byte
	var found bool
	err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b != nil {
			if v := b.Get(p); v != nil {
				data = make([]byte, len(v))
				copy(data, v)
				found = true
			}
		}
		return nil
	})
	return data, found, err
}

func (l *boltDBKV) Close() error {
	return l.db.Close()
}
