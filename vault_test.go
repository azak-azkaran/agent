package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultIsSealed(t *testing.T) {
	fmt.Println("running: TestVaultGetStatus")
	testconfig := readConfig(t)

	resp, err := IsSealed(testconfig.config)
	assert.NoError(t, err)
	assert.False(t, resp, ERROR_VAULT_SEALED)
}

func TestVaultGetSecret(t *testing.T) {
	fmt.Println("running: TestVaultGetSecret")
	testconfig := readConfig(t)
	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal, ERROR_VAULT_SEALED)

	resp, err := GetSecret(testconfig.config, testconfig.token, "restic/data/"+testconfig.resticpath)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Data["data"])
}
func TestVaultUnseal(t *testing.T) {
	fmt.Println("running: TestVaultUnseal")

	testconfig := readConfig(t)
	err := Seal(testconfig.config, VAULT_TEST_TOKEN)
	require.NoError(t, err)

	seal, err := IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.True(t, seal)

	_, err = Unseal(testconfig.config, testconfig.secret)
	require.NoError(t, err)

	seal, err = IsSealed(testconfig.config)
	require.NoError(t, err)
	assert.False(t, seal)
}

func TestVaultSealStatus(t *testing.T) {
	fmt.Println("running: TestVaultSealStatus")
	testconfig := readConfig(t)
	multipleKey = true
	err := Seal(testconfig.config, VAULT_TEST_TOKEN)
	require.NoError(t, err)

	//testconfig.config.Address = "http://127.0.0.1:8200"

	seal, err := SealStatus(testconfig.config)
	assert.NoError(t, err)

	assert.Equal(t, 5, seal.N)
	assert.Equal(t, 3, seal.T)
	assert.Equal(t, 0, seal.Progress)

	assert.True(t, seal.Sealed)

	seal, err = Unseal(testconfig.config, VAULT_TEST_TOKEN)
	assert.NoError(t, err)
	assert.Equal(t, 1, seal.Progress)

	seal, err = SealStatus(testconfig.config)
	assert.NoError(t, err)
	assert.Equal(t, 1, seal.Progress)

	seal, err = Unseal(testconfig.config, VAULT_TEST_TOKEN)
	assert.NoError(t, err)
	assert.Equal(t, 2, seal.Progress)

	seal, err = Unseal(testconfig.config, VAULT_TEST_TOKEN)
	assert.NoError(t, err)
	assert.Equal(t, 0, seal.Progress)
	assert.False(t, seal.Sealed)

	seal, err = SealStatus(testconfig.config)
	assert.NoError(t, err)
	assert.False(t, seal.Sealed)

	multipleKey = false
}
