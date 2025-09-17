package main

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

func main() {

	db, err := badger.Open(badger.DefaultOptions("./testdata/badger-demo"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("key"), []byte("value"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Println("set error ", err)
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	// Use the transaction...
	err = txn.Set([]byte("answer"), []byte("42"))
	if err != nil {
		return
	}

	// Commit the transaction and check for error.
	if err = txn.Commit(); err != nil {
		return
	}

	db.View(func(txn *badger.Txn) error {
		item, rerr := txn.Get([]byte("key"))
		if rerr != nil {
			return rerr
		}
		fmt.Println(item.String())
		return nil
	})

}
