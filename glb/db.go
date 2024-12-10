package glb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"ju"
)

var servDb *leveldb.DB

func init() {
	var err error
	servDb, err = leveldb.OpenFile("./data/db", nil)
	ju.CheckError(err)
}
func GetDb() *leveldb.DB {
	return servDb
}
