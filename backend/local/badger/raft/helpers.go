package badgerraftcache

import (
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

func (c *Cache) FirstIndex() (uint64, error) {
	tx := c.db.NewTransaction(false)
	defer tx.Discard()
	iter := tx.NewIterator(iterAscOpt)
	iter.Rewind()
	item := iter.Item()
	if item == nil {
		return 0, nil
	}

	return bytesToUint64(item.Key()), nil
}

func (c *Cache) LastIndex() (uint64, error) {
	tx := c.db.NewTransaction(false)
	defer tx.Discard()
	iter := tx.NewIterator(iterDescOpt)
	iter.Rewind()
	item := iter.Item()
	if item == nil {
		return 0, nil
	}
	return bytesToUint64(item.Key()), nil
}

func (c *Cache) DeleteRange(min, max uint64) error {
	tx := c.db.NewTransaction(true)
	defer tx.Discard()
	minKey := uint64ToBytes(min)
	iter := tx.NewIterator(iterAscOpt)
	for iter.Seek(minKey); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item == nil {
			break
		}
		curKey := safeKey(item)
		if bytesToUint64(curKey) > max {
			break
		}
		if err := tx.Delete(curKey); err != nil {
			return err
		}
	}
	if err := tx.Commit(nil); err != nil {
		return err
	}
	return nil
}
