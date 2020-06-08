package main

import (
	"log"

	badger "github.com/dgraph-io/badger/v2"
)

func InitDB(path string, debug bool) *badger.DB {
	var opt badger.Options
	if debug {
		log.Println("Debug is on switching to InMemory")

		opt = badger.DefaultOptions("").WithInMemory(true)
	} else {
		opt = badger.DefaultOptions(path)
	}

	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func Get(db *badger.DB, key string) (string, error) {
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		// Alternatively, you could also use item.ValueCopy().
		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return string(valCopy), nil
}

func Put(db *badger.DB, key string, value string) (bool, error) {
	var ok bool
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte(value))
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
		//err = txn.Commit()
		//if err != nil {
		//	return err
		//}
		ok = true
		return nil
	})

	return ok, err
}
