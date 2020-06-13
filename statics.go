package main

const (
	// Backup Constants
	RESTIC_PASSWORD   = "RESTIC_PASSWORD="
	RESTIC_REPOSITORY = "RESTIC_REPOSITORY="
	RESTIC_ACCESS_KEY = "AWS_ACCESS_KEY_ID="
	RESTIC_SECRET_KEY = "AWS_SECRET_ACCESS_KEY="

	// Store Constants
	STORE_TOKEN     = "token"
	STORE_TIMESTAMP = "timestamp"

	MAIN_PATHDB         = "pathdb"
	MAIN_ADDRESS        = "address"
	MAIN_TIME_DURATION  = "duration"
	MAIN_MOUNT_DURATION = "mount-duration"
	MAIN_MOUNT_ALLOW    = "mount-allow"

	ERROR_DATABASE_NOT_FOUND = "Database is not initialized"

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
	ERROR_PUT_TOKEN = "PutToken:"
	ERROR_PREFIX    = "ERROR: "
	JSON_MESSAGE    = "message"

	ERROR_VAULT_SEALED    = "Vault is sealed."
	ERROR_VAULT_NO_SECRET = "Vault has no data for this endpoint."

	ERROR_UNMARSHAL        = "Error marshaling message: "
	ERROR_SENDING_REQUEST  = "Error sending request: "
	ERROR_READING_RESPONSE = "Error reading response: "

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

	MAIN_TEST_ADDRESS        = "localhost:8031"
	MAIN_TEST_PATHDB         = "./test/DB"
	MAIN_TEST_DURATION       = "1h30m"
	MAIN_TEST_MOUNT_DURATION = "5s"
	MAIN_TEST_MOUNT_ALLOW    = "false"

	REST_TEST_TOKEN     = "http://localhost:8031/token"
	REST_TEST_LOG       = "http://localhost:8031/logs"
	REST_TEST_BACKUP    = "http://localhost:8031/backup"
	REST_TEST_STATUS    = "http://localhost:8031/status"
	REST_TEST_MOUNT     = "http://localhost:8031/mount"
	REST_TEST_UNSEAL    = "http://localhost:8031/unseal"
	REST_TEST_IS_SEALED = "http://localhost:8031/is_sealed"
	REST_TEST_SEAL      = "http://localhost:8031/seal"
	REST_TEST_PING      = "http://localhost:8031/ping"
)
