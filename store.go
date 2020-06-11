package main

import (
	"log"
	"time"

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
		log.Println("Error opening database: ", err)
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

func UpdateTimestamp(db *badger.DB, timestamp time.Time) (bool, error) {
	return Put(db, STORE_TIMESTAMP, timestamp.Format(time.RFC3339Nano))
}

func GetTimestamp(db *badger.DB) (time.Time, error) {
	value, err := Get(db, STORE_TIMESTAMP)
	if err != nil {
		return time.Unix(0, 0), err
	}

	n, err := time.Parse(time.RFC3339Nano, value)

	//n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}

	return n, nil
}

func CheckToken(db *badger.DB) bool {
	value, err := Get(db, STORE_TOKEN)
	if err != nil {
		return false
	}

	if value == "" {
		return false
	}

	return true
}

func GetToken(db *badger.DB) (string, error) {
	return Get(db, STORE_TOKEN)
}

func PutToken(db *badger.DB, token string) (bool, error) {
	return Put(db, STORE_TOKEN, token)
}
