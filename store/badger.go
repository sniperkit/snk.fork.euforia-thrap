package store

import (
	"bytes"
	"errors"
	"hash"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/euforia/thrap/thrapb"
)

var (
	errPrevHashMismatch = errors.New("previous hash mismatch")
	errObjectExists     = errors.New("object exists")
	errRefExists        = errors.New("ref exists")
)

const (
	objKeyPrefix = "objects/"
	refKeyPrefix = "refs/"
)

// Object is a datastructure that is hashable and protobuf friendly
type Object interface {
	Hash(h hash.Hash) []byte
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
}

func NewBadgerDB(datadir string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = datadir
	opts.ValueDir = datadir
	opts.SyncWrites = true
	return badger.Open(opts)
}

type BadgerObjectStore struct {
	db         *badger.DB
	prefix     string
	hf         func() hash.Hash
	zeroDigest []byte // zero digest based on hash function
}

// NewBadgerObjectStore returns a badger backed manifest store
func NewBadgerObjectStore(db *badger.DB, hf func() hash.Hash, prefix string) *BadgerObjectStore {
	b := &BadgerObjectStore{
		db:         db,
		prefix:     prefix,
		hf:         hf,
		zeroDigest: make([]byte, hf().Size()),
	}
	if b.prefix[len(b.prefix)-1] != '/' {
		b.prefix += "/"
	}

	return b
}

func (store *BadgerObjectStore) CreateRef(namespace, ref string) ([]byte, *thrapb.ChainHeader, error) {
	h := store.hf()
	ch := &thrapb.ChainHeader{
		Previous:   make([]byte, h.Size()),
		Timestamp:  time.Now().UnixNano(),
		DataDigest: make([]byte, h.Size()),
	}
	b, _ := ch.Marshal()
	digest := ch.Hash(h)

	refkey := store.getRefOpaque(namespace, ref)

	er := store.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(refkey); err == nil {
			return errRefExists
		}
		err := txn.Set(store.getObjOpaque(namespace, digest), b)
		if err == nil {
			err = txn.Set(refkey, digest)
		}
		return err
	})

	return digest, ch, er
}

func (store *BadgerObjectStore) IterRefChain(namespace, ref string, f func(header *thrapb.ChainHeader) error) error {
	refkey := store.getRefOpaque(namespace, ref)
	er := store.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(refkey)
		if err != nil {
			return err
		}
		digest, err := item.Value()
		if err != nil {
			return err
		}

		for err == nil && !store.isZeroDigest(digest) {
			var robj *thrapb.ChainHeader
			robj, err = store.getRefObj(txn, namespace, digest)
			if err == nil {
				if err = f(robj); err == nil {
					digest = robj.Previous
				}
			}

		}
		return err
	})

	return er
}

func (store *BadgerObjectStore) isZeroDigest(in []byte) bool {
	return bytes.Compare(store.zeroDigest, in) == 0
}

func (store *BadgerObjectStore) SetRef(namespace, ref string, ch *thrapb.ChainHeader) ([]byte, error) {

	refkey := store.getRefOpaque(namespace, ref)
	h := store.hf()

	var digest []byte
	err := store.db.Update(func(txn *badger.Txn) error {
		// Check specified previous exists
		refobj, er := store.getRefObj(txn, namespace, ch.Previous)
		if er != nil {
			return er
		}

		ch.Height = refobj.Height + 1
		b, er := ch.Marshal()
		if er != nil {
			return er
		}

		digest = ch.Hash(h)
		objkey := store.getObjOpaque(namespace, digest)

		er = txn.Set(objkey, b)
		if er == nil {
			er = txn.Set(refkey, digest)
		}
		return er
	})

	return digest, err
}

func (store *BadgerObjectStore) GetRef(namespace, ref string) (*thrapb.ChainHeader, []byte, error) {
	var (
		//refkey = store.getRefOpaque(namespace, ref)
		digest []byte
		pb     *thrapb.ChainHeader
	)

	er := store.db.View(func(txn *badger.Txn) error {
		var err error
		pb, digest, err = store.getRef(txn, namespace, ref)
		return err
	})

	return pb, digest, er
}

func (store *BadgerObjectStore) getRef(txn *badger.Txn, namespace, ref string) (*thrapb.ChainHeader, []byte, error) {
	refkey := store.getRefOpaque(namespace, ref)
	item, err := txn.Get(refkey)
	if err != nil {
		return nil, nil, err
	}
	digest, err := item.Value()
	if err != nil {
		return nil, nil, err
	}

	header, err := store.getRefObj(txn, namespace, digest)
	return header, digest, err
}

func (store *BadgerObjectStore) DeleteRef(namespace, ref string) error {
	key := store.getRefOpaque(namespace, ref)
	return store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (store *BadgerObjectStore) Get(namespace string, digest []byte, out Object) error {
	key := store.getObjOpaque(namespace, digest)

	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err := item.Value()
		if err == nil {
			err = out.Unmarshal(val)
		}

		return err
	})

	return err
}

func (store *BadgerObjectStore) Set(namespace string, obj Object) ([]byte, error) {
	h := store.hf()
	digest := obj.Hash(h)
	objkey := store.getObjOpaque(namespace, digest)

	err := store.db.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get(objkey); err == nil {
			return errObjectExists
		}

		val, err := obj.Marshal()
		if err == nil {
			err = txn.Set(objkey, val)
		}

		return err
	})

	return digest, err
}

// Delete deletes an object by the given digest
func (store *BadgerObjectStore) Delete(namespace string, digest []byte) error {
	return store.db.Update(func(txn *badger.Txn) error {
		key := store.getObjOpaque(namespace, digest)
		return txn.Delete(key)
	})
}

func (store *BadgerObjectStore) getRefOpaque(namespace, ref string) []byte {
	str := filepath.Join(store.prefix, namespace, refKeyPrefix, ref)
	return []byte(str)
	// return append(store.prefix, append([]byte(refKeyPrefix), key...)...)
}

func (store *BadgerObjectStore) getObjOpaque(namespace string, digest []byte) []byte {
	str := filepath.Join(store.prefix, namespace, objKeyPrefix)
	return append(append([]byte(str), byte('/')), digest...)
	// return append(store.prefix, append([]byte(objKeyPrefix), digest...)...)
}

func (store *BadgerObjectStore) getRefObj(txn *badger.Txn, namespace string, digest []byte) (*thrapb.ChainHeader, error) {
	rkey := store.getObjOpaque(namespace, digest)
	item, err := txn.Get(rkey)
	if err != nil {
		return nil, err
	}

	pb := new(thrapb.ChainHeader)

	val, err := item.Value()
	if err == nil {
		err = pb.Unmarshal(val)
	}

	return pb, err
}
