package glb

import (
	"github.com/jsuserapp/ju"
	"github.com/syndtr/goleveldb/leveldb"
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
