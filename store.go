package main

import (
	"encoding/binary"
	"errors"
	"strconv"
	"time"

	crypto_rand "crypto/rand"
	math_rand "math/rand"

	badger "github.com/dgraph-io/badger/v3"
)

var closed = true

type defaultLog struct{}

func (l *defaultLog) Errorf(f string, v ...interface{}) {
	Sugar.Errorf(f, v...)
}

func (l *defaultLog) Infof(f string, v ...interface{}) {
	Sugar.Infof(f, v...)
}

func (l *defaultLog) Debugf(f string, v ...interface{}) {
	Sugar.Debugf(f, v...)
}
func (l *defaultLog) Warningf(f string, v ...interface{}) {
	Sugar.Warnf(f, v...)
}

func InitDB(path string, masterkey string, debug bool) *badger.DB {
	var opt badger.Options

	if debug {
		Sugar.Warn("Debug is on switching to InMemory")

		opt = badger.DefaultOptions("").WithInMemory(true).WithLogger(&defaultLog{})
	} else {
		opt = badger.DefaultOptions(path).WithLogger(&defaultLog{})
	}

	if masterkey != "" {
		opt.WithEncryptionKey([]byte(masterkey))
	}

	db, err := badger.Open(opt)
	if err != nil {
		Sugar.Error("Error opening database: ", err)
	}
	closed = false
	return db
}

func Get(db *badger.DB, key string) (string, error) {
	if db == nil {
		return "", errors.New(ERROR_DATABASE_NOT_FOUND)
	}
	if closed {
		return "", errors.New(ERROR_DATABASE_CLOSED)
	}
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
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
	if db == nil {
		return false, errors.New(ERROR_DATABASE_NOT_FOUND)
	}
	if closed {
		return false, errors.New(ERROR_DATABASE_CLOSED)
	}
	var ok bool
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte(value))
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
		ok = true
		return nil
	})

	return ok, err
}

func Remove(db *badger.DB, key string) ( error) {
	if db == nil {
		return  errors.New(ERROR_DATABASE_NOT_FOUND)
	}
	if closed {
		return  errors.New(ERROR_DATABASE_CLOSED)
	}
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return  err
	}
	return  nil
}

func UpdateLastBackup(db *badger.DB, timestamp time.Time) (bool, error) {
	return Put(db, STORE_LAST_BACKUP, timestamp.Format(time.RFC3339Nano))
}

func GetLastBackup(db *badger.DB) (time.Time, error) {
	return getTimestamp(db, STORE_LAST_BACKUP)
}

func UpdateTimestamp(db *badger.DB, timestamp time.Time) (bool, error) {
	return Put(db, STORE_TIMESTAMP, timestamp.Format(time.RFC3339Nano))
}

func GetTimestamp(db *badger.DB) (time.Time, error) {
	return getTimestamp(db, STORE_TIMESTAMP)
}

func getTimestamp(db *badger.DB, key string) (time.Time, error) {
	value, err := Get(db, key)
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

func CheckSealKey(db *badger.DB, shares int) bool {
	for i := 1; i < shares+1; i++ {
		value, err := Get(db, STORE_KEY+strconv.Itoa(shares))
		if err != nil {
			return false
		}

		if value == "" {
			return false
		}
	}
	return true
}

func PutSealKey(db *badger.DB, key string, shares int) (bool, error) {
	Sugar.Info("Adding seal key, ", shares)
	return Put(db, STORE_KEY+strconv.Itoa(shares), key)
}

func DropSealKeys(db *badger.DB, length int) error {
	if db == nil {
		return errors.New(ERROR_DATABASE_NOT_FOUND)
	}
	if closed {
		return errors.New(ERROR_DATABASE_CLOSED)
	}
	if err := db.Sync(); err != nil {
		return err
	}

	err := db.DropPrefix([]byte(STORE_KEY))
	if err != nil {
		return err
	}

	ok := CheckSealKey(db, 1)
	if ok {
		return errors.New(STORE_ERROR_NOT_DROPED)
	}
	return nil
}

func GetSealKey(db *badger.DB, threshold int, totalShares int) []string {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	var values []string
	if err != nil {
		Sugar.Error("cannot seed math/rand package with cryptographically secure random number generator")
		return values
	}

	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))

	permutation := math_rand.Perm(totalShares)
	for i := 0; i < threshold; i++ {
		v := permutation[i] + 1
		value, err := Get(db, STORE_KEY+strconv.Itoa(v))
		if err != nil {
			Sugar.Error(ERROR_KEY_NOT_FOUND, err)
		} else {
			values = append(values, value)
		}
	}
	return values
}

func Close(db *badger.DB, timeout time.Duration) {
	closed = true
	time.Sleep(timeout)
	db.Close()
}
