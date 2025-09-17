package main

import (
	"fmt"

	badger4 "github.com/dgraph-io/badger/v4"
)

func main() {

	db, err := badger4.Open(badger4.DefaultOptions("./testdata/badger4-demo"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	db.Update(func(txn *badger4.Txn) error {

		if err := txn.Set([]byte("answer"), []byte("42")); err != nil {
			return err
		}
		return nil
	})

	txn := db.NewTransaction(true)
	item, err := txn.Get([]byte("answer"))
	if err != nil {
		fmt.Println(err)
		return
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(val))

}
