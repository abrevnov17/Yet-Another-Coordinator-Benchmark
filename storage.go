package main

import (
	"sync"

	"github.com/peterbourgon/diskv"
)

var sagasMutex = &sync.Mutex{}
var sagas = make(map[string]Saga)

var db = diskv.New(diskv.Options{
	BasePath:     "diskv",
	CacheSizeMax: 0,
})

func writeToDisk(key string, value *Saga) {
	arr := value.toByteArray()
	_ = db.Write(key, arr)
}

func readFromDisk(key string) *Saga {
	value, _ := db.Read(key)
	return fromByteArray(value)
}

func removeFromDisk(key string) {
	_ = db.Erase(key)
}

func updateDisk(key, partial string, value Request) {
	s := readFromDisk(key)
	s.PartialReqs[partial] = value
	writeToDisk(key, s)
}
