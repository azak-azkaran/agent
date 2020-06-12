package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreInitDB(t *testing.T) {
	fmt.Println("running: TestStoreInitDB")

	db := InitDB("", true)
	require.NotNil(t, db)

	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte("answer"), []byte("42"))
		err := txn.SetEntry(e)
		return err // Your code here…
	})
	assert.NoError(t, err)

	var valCopy []byte
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("answer"))
		require.NoError(t, err)

		// Alternatively, you could also use item.ValueCopy().
		valCopy, err = item.ValueCopy(nil)
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	test, err := strconv.Atoi(string(valCopy))
	require.NoError(t, err)
	assert.Equal(t, 42, test)

}

func TestStoreIntegration(t *testing.T) {
	fmt.Println("running: TestStoreIntegration")

	db := InitDB("", true)
	require.NotNil(t, db)

	ok, err := Put(db, "answer", "42")
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err := Get(db, "answer")
	assert.NoError(t, err)
	assert.Equal(t, "42", value)

	ok, err = Put(db, "answer", "theAnswer")
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err = Get(db, "answer")
	assert.NoError(t, err)
	assert.Equal(t, "theAnswer", value)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")
}

func TestStorePut(t *testing.T) {
	fmt.Println("running: TestStorePut")
	db := InitDB("", true)
	require.NotNil(t, db)

	ok, err := Put(db, "answer", "42")
	assert.NoError(t, err)
	assert.True(t, ok)

	var valCopy []byte
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("answer"))
		require.NoError(t, err)

		// Alternatively, you could also use item.ValueCopy().
		valCopy, err = item.ValueCopy(nil)
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	test, err := strconv.Atoi(string(valCopy))
	require.NoError(t, err)
	assert.Equal(t, 42, test)

}

func TestStoreGet(t *testing.T) {
	fmt.Println("running: TestStoreGet")
	db := InitDB("", true)
	require.NotNil(t, db)

	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte("answer"), []byte("42"))
		err := txn.SetEntry(e)
		return err // Your code here…
	})
	assert.NoError(t, err)

	val, err := Get(db, "answer")
	assert.NoError(t, err)

	test, err := strconv.Atoi(string(val))
	require.NoError(t, err)
	assert.Equal(t, 42, test)
}

func TestStoreUpdateTimestamp(t *testing.T) {
	fmt.Println("running: TestStoreUpdateTimestamp")
	db := InitDB("", true)
	require.NotNil(t, db)

	value, err := GetTimestamp(db)
	assert.Error(t, err)
	assert.Equal(t, time.Unix(0, 0), value)

	timestamp := time.Now()

	ok, err := UpdateTimestamp(db, timestamp)
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err = GetTimestamp(db)
	assert.NoError(t, err)
	assert.Equal(t, timestamp.Unix(), value.Unix())
	assert.Equal(t, timestamp.Format(time.RFC3339Nano), value.Format(time.RFC3339Nano))
}

func TestStoreCheckToken(t *testing.T) {
	fmt.Println("running: TestStoreCheckToken")
	db := InitDB("", true)
	require.NotNil(t, db)

	ok := CheckToken(db)
	assert.False(t, ok)

	ok, err := PutToken(db, "testToken")
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err := GetToken(db)
	assert.NoError(t, err)
	assert.Equal(t, "testToken", value)

	ok = CheckToken(db)
	assert.True(t, ok)

}
