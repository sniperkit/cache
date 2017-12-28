package badgerraftcache

import (
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger"
	// "github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
)

const (
	defaultCacheValueDir string = "httpcache"
	defaultCacheDir      string = "./shared/data/cache/.badger"
)

var defaultCachePath string = filepath.Join(defaultCacheDir, defaultCacheValueDir)

/*
	Refs:
	- https://github.com/dtynn/raftbadger/blob/master/badger_store.go
*/

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	sync.RWMutex
	db          *badger.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	StoragePath string
	ValueDir    string
	// SyncWrites  bool
	Logger        bool
	LogsForwarder bool
	AccessLogs    string
	ErrorLogs     string
	Debug         bool
}

func Mount(client *badger.DB) *Cache {
	return &Cache{db: client}
}

func New(config *Config) (*Cache, error) {
	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("badgerraftcache.New()")
	}
	badgerConfig := badger.DefaultOptions
	if config == nil {
		badgerConfig.Dir = defaultCacheDir
		badgerConfig.ValueDir = defaultCacheValueDir
		// badgerConfig.SyncWrites = false
	} else {
		badgerConfig.Dir = config.StoragePath
		badgerConfig.ValueDir = config.ValueDir
		// badgerConfig.SyncWrites = config.SyncWrites
	}
	client, err := badger.Open(badgerConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
		}).Fatalln("badgerraftcache.New().badger.Open(), ERROR: ", err)
		return nil, err
	}
	// defer db.Close()
	return &Cache{
		db: client,
	}, nil
}

func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.Lock()
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		resp, err := item.Value()
		if err != nil {
			return err
		}
		if c.debug {
			log.WithFields(log.Fields{
				"resp": resp,
			}).Debug("badgercache.Get()")
		}
		return nil
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debug("badgercache.Get()")
	}
	return resp, err != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(resp))
		return err
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debugln("badgercache.Set()")
	}
	return
}

func (c *Cache) Delete(key string) {
	c.Lock()
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debug("badgercache.Delete()")
	}
	return
}

func (c *Cache) SetUint64(key []byte, val uint64) error {
	return c.Set(key, uint64ToBytes(val))
}

func (c *Cache) GetUint64(key []byte) (uint64, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

/*
func (c *Cache) Set(key, val []byte) error {
	tx := c.db.NewTransaction(true)
	defer tx.Discard()
	if err := tx.Set(key, val); err != nil {
		return err
	}
	if err := tx.Commit(nil); err != nil {
		return err
	}
	return nil
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	tx := c.db.NewTransaction(false)
	defer tx.Discard()
	item, err := tx.Get(key)
	if err != nil {
		return nil, err
	}
	return item.ValueCopy(nil)
}

*/

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}
