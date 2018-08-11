/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package store

import (
	"github.com/dgraph-io/badger"
)

// NewBadgerDB opens a new badger db handle from the given directory
func NewBadgerDB(datadir string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = datadir
	opts.ValueDir = datadir
	opts.SyncWrites = true
	return badger.Open(opts)
}
