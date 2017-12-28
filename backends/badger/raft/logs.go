package badgerraftcache

import (
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

func (c *Cache) GetLog(idx uint64, log *raft.Log) error {
	tx := c.db.NewTransaction(false)
	defer tx.Discard()
	item, err := tx.Get(uint64ToBytes(idx))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return raft.ErrLogNotFound
		}
		return err
	}
	val, err := item.Value()
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"idx":    idx,
		}).Fataln("badgerraftcache.GetLog().item.Value(), ERROR:", err)
		return err
	}
	return decodeMsgPack(val, log)
}

func (c *Cache) StoreLog(log *raft.Log) error {
	return c.StoreLogs([]*raft.Log{log})
}

func (c *Cache) StoreLogs(logs []*raft.Log) error {
	tx := b.logdb.NewTransaction(true)
	defer tx.Discard()
	for _, one := range logs {
		buf, err := encodeMsgPack(one)
		if err != nil {
			return err
		}
		if err := tx.Set(uint64ToBytes(one.Index), buf.Bytes()); err != nil {
			return err
		}
	}
	if err := tx.Commit(nil); err != nil {
		log.WithFields(log.Fields{
			"config": config,
		}).Fataln("badgerraftcache.StoreLogs().Commit, ERROR:", err)
		return err
	}
	return nil
}
