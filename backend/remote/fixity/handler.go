package fixtycache

import (
	"fmt"
	"sync"
	"time"
	// "path/filepath"
	// "errors"

	"github.com/leeola/fixity"
	"github.com/leeola/fixity/autoload"

	log "github.com/sirupsen/logrus"
)

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.fixity"
	defaultStorageFileExtension string = ".fixity"
)

var (
	defaultStorageDir  string = filepath.Join(bktDir, bktName)
	defaultStorageFile string = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

// Cache is an implementation of httpcache.Cache that uses a bolt database.
type Cache struct {
	sync.RWMutex
	db          *fixity.DB
	storagePath string
	bucketName  string
	debug       bool
	stats       bool
	compress    bool
}

type Config struct {
	BucketName     string
	StoragePath    string
	ReadOnly       bool
	StrictMode     bool
	NoSync         bool
	NoFreelistSync bool
	NoGrowSync     bool
	MaxBatchSize   bool
	MaxBatchDelay  bool
	AllocSize      bool
	Compress       bool
	Debug          bool
	Stats          bool
}

type Check struct {
	Enabled   bool
	Key       string
	Requests  int
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
	Priority  bool
	Provider  string
}

// New returns a new Cache that uses a bolt database at the given path.
func New(config *Config) (*Cache, error) {

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("fixitycache.New()")
	}

	if config == nil {
		log.Warnln("config is nil")
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStorageFile
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("fixitycache.New() ---> post-processed")
	}

	if config == nil {
		config.StoragePath = defaultStoragePath
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStoragePath
	}

	db, err := autoload.LoadFixity(config.StoragePath)
	if err != nil {
		fmt.Println("Error occured while loading fixity")
	}
	return &Cache{
		store: db,
	}, nil
}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *fixity.DB) *Cache {
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
		}).Warnf("fixitycache.Close()")
	}
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.RLock()
	c.RUnlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
		}).Info("fixitycache.Get() OK")
	}
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
		}).Info("fixitycache.Set() OK")
	}
	return
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	c.Lock()
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
		}).Info("fixitycache.Delete() OK")
	}
	return
}

func (c *Cache) Debug(action string) {}

func (c *Cache) Action(name string, args ...interface{}) (map[string]*interface{}, error) {
	return nil, errors.New("Action not implemented yet")
}
