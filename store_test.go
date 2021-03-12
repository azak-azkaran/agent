package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreInitDB(t *testing.T) {
	fmt.Println("running: TestStoreInitDB")

	db := InitDB("./test/DB", "", false)
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

	err = db.Close()
	assert.NoError(t, err)

	fmt.Println("Testing with Masterkey")
	db = InitDB("./test/DB", "test", false)
	require.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}

func TestStoreIntegration(t *testing.T) {
	fmt.Println("running: TestStoreIntegration")

	db := InitDB("./test/DB", "", false)
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

	err = db.Close()
	assert.NoError(t, err)

	fmt.Println("Testing with Masterkey")
	db = InitDB("./test/DB", "test", false)
	require.NotNil(t, db)

	ok, err = Put(db, "answer", "Blub")
	assert.NoError(t, err)
	assert.True(t, ok)

	err = db.Close()
	assert.NoError(t, err)

	err = RemoveContents("./test/DB/")
	assert.NoError(t, err)
	assert.NoFileExists(t, "./test/DB/MANIFEST")

}

func TestStorePut(t *testing.T) {
	fmt.Println("running: TestStorePut")
	db := InitDB("", "", true)
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

	err = db.Close()
	assert.NoError(t, err)
}

func TestStoreGet(t *testing.T) {
	fmt.Println("running: TestStoreGet")
	db := InitDB("", "", true)
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

	err = db.Close()
	assert.NoError(t, err)
}

func TestStoreUpdateTimestamp(t *testing.T) {
	fmt.Println("running: TestStoreUpdateTimestamp")
	db := InitDB("", "", true)
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

	err = db.Close()
	assert.NoError(t, err)
}

func TestStoreCheckToken(t *testing.T) {
	fmt.Println("running: TestStoreCheckToken")
	db := InitDB("", "", true)
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

	err = db.Close()
	assert.NoError(t, err)
}

func TestStoreCheckSealKey(t *testing.T) {
	fmt.Println("running: TestStoreGetSealKey")
	db := InitDB("", "", true)
	require.NotNil(t, db)

	ok := CheckSealKey(db, 1)
	assert.False(t, ok)

	key := "test"
	for i := 1; i < 6; i++ {
		ok, err := PutSealKey(db, key, i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	ok = CheckSealKey(db, 1)
	assert.True(t, ok)

	ok = CheckSealKey(db, 5)
	assert.True(t, ok)

	ok = CheckSealKey(db, 6)
	assert.False(t, ok)

	err := db.Close()
	assert.NoError(t, err)
}

func TestStoreGetSealKey(t *testing.T) {
	fmt.Println("running: TestStoreGetSealKey")
	db := InitDB("", "", true)
	require.NotNil(t, db)

	key := "test"
	for i := 1; i < 6; i++ {
		ok, err := PutSealKey(db, key, i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	values := GetSealKey(db, 3, 5)
	assert.Len(t, values, 3)

	err := db.Close()
	assert.NoError(t, err)

}

func TestStoreDropSealKeys(t *testing.T) {
	fmt.Println("running: TestStoreDropSealKeys")
	db := InitDB("", "", true)
	require.NotNil(t, db)

	ok, err := PutSealKey(db, "test", 1)
	assert.NoError(t, err)
	assert.True(t, ok)

	keys := GetSealKey(db, 1, 1)
	assert.NotNil(t, keys)
	assert.Len(t, keys, 1)

	err = DropSealKeys(db)
	assert.NoError(t, err)

	key, err := Get(db, STORE_KEY+"1")
	assert.Error(t, err)
	assert.Equal(t, "", key)
	keys = GetSealKey(db, 1, 1)
	assert.Len(t, keys, 0)
}
