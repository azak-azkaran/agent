package main

const (
	// Backup Constants
	RESTIC_PASSWORD   = "RESTIC_PASSWORD="
	RESTIC_REPOSITORY = "RESTIC_REPOSITORY="
	RESTIC_ACCESS_KEY = "AWS_ACCESS_KEY_ID="
	RESTIC_SECRET_KEY = "AWS_SECRET_ACCESS_KEY="

	// Restrouter
	ERROR_MODE      = "Backup Mode:"
	ERROR_STATUS    = "GetStatus:"
	ERROR_LOG       = "GetLogs:"
	ERROR_ISSEALED  = "IsSealed:"
	ERROR_UNSEAL    = "Unseal:"
	ERROR_SEAL      = "Seal:"
	ERROR_RUNBACKUP = "RunBackupJob:"
	ERROR_RUNMOUNT  = "RunMountJob:"
	ERROR_ENQUEUE   = "Enqueue:"
	ERROR_CONFIG    = "GetConfigFromVault:"
	ERROR_BINDING   = "BindJSON:"
	ERROR_PREFIX    = "ERROR: "
	JSON_MESSAGE    = "message"

	ERROR_VAULT_SEALED    = "Vault is sealed."
	ERROR_VAULT_NO_SECRET = "Vault has no data for this endpoint."

	BACKUP_TEST_FOLDER       = "./test/Backup"
	BACKUP_TEST_EXCLUDE_FILE = "./test/exclude"
	BACKUP_TEST_CONF_FILE    = "./test/Backup/config"

	// Mount Constants For Tests
	GOCRYPT_TEST_MOUNTPATH = "./test/tmp-mount"
	GOCRYPT_TEST_FILE      = "./test/tmp-mount/test"
	GOCRYPT_TEST_FOLDER    = "./test/tmp"

	VAULT_TEST_PASSWORD            = "hallo"
	VAULT_TEST_TOKEN               = "superrandompasswordtoken"
	VAULT_TEST_PATH                = "./test/tmp"
	VAULT_TEST_MOUNTPATH           = "./test/tmp-mount"
	VAULT_TEST_CONFIGPATH          = "gocryptpath"
	VAULT_TEST_BACKUP_PATH         = "./test/Backup"
	VAULT_TEST_BACKUP_EXCLUDE_FILE = "./test/exclude"
	VAULT_TEST_BACKUP_SECRET_KEY   = "secret.key"
	VAULT_TEST_BACKUP_ACCESS_KEY   = "access.key"

	MAIN_TEST_ADDRESS = "localhost:8081"
)
