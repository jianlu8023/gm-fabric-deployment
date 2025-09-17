package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// https://github.com/syndtr/goleveldb
func main() {

	db, err := leveldb.OpenFile("./testdata/leveldb-demo", &opt.Options{})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	err = db.Put([]byte("key"), []byte("value"), nil)
	data, err := db.Get([]byte("key"), nil)
	fmt.Println(string(data))
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("key=%s, value=%s\n", key, value)
	}
	iter.Release()
	err = iter.Error()

	err = db.Delete([]byte("key"), nil)

}
