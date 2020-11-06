package main

const (
	// Backup Constants
	RESTIC_PASSWORD   = "RESTIC_PASSWORD="
	RESTIC_REPOSITORY = "RESTIC_REPOSITORY="
	RESTIC_ACCESS_KEY = "AWS_ACCESS_KEY_ID="
	RESTIC_SECRET_KEY = "AWS_SECRET_ACCESS_KEY="

	// Git Contstatns
	GIT_REMOTE_NAME = "agent_remote"

	// Store Constants
	STORE_TOKEN       = "token"
	STORE_TIMESTAMP   = "timestamp"
	STORE_LAST_BACKUP = "last_backup"
	STORE_KEY         = "vault-key-"

	STORE_ERROR_NOT_DROPED = "Error keys were not dropped."

	MAIN_PATHDB          = "pathdb"
	MAIN_ADDRESS         = "address"
	MAIN_TIME_DURATION   = "duration"
	MAIN_MOUNT_DURATION  = "mount_duration"
	MAIN_MOUNT_ALLOW     = "mount_allow"
	MAIN_VAULT_KEY_FILE  = "vault_key_file"
	MAIN_VAULT_ADDRESS   = "vault_address"
	MAIN_VAULT_SECRET_ID = "vault_secret_id"
	MAIN_VAULT_ROLE_ID   = "vault_role_id"

	MAIN_MESSAGE_NOT_ENOUGH_KEYS  = "Not enough vault keys in storage"
	MAIN_MESSAGE_START_UNSEAL     = "Starting to unseal Vault"
	MAIN_MESSAGE_START_RESTSERVER = "Starting the REST Server"
	MAIN_MESSAGE_START_RUNNING    = "Starting the Agent RUN - Function in 5 Seconds"
	MAIN_MESSAGE_BACKUP_INIT      = "Backup Repository not found will initialize it"
	MAIN_MESSAGE_BACKUP_SUCCESS   = "Backup Success"
	MAIN_MESSAGE_BACKUP_ALREADY   = "Backup Check was already run at: "

	MAIN_POST_HTTP            = "http://"
	MAIN_POST_BACKUP_ENDPOINT = "/backup"
	MAIN_POST_MOUNT_ENDPOINT  = "/mount"
	MAIN_POST_GIT_ENDPOINT    = "/git"
	MAIN_POST_DATA_TYPE       = "application/json"

	MAIN_ERROR_CHECK_SEAL = "Error while checking seal: "
	MAIN_ERROR_UNSEAL     = "Error while unsealing vault: "
	MAIN_ERROR_SHUTDOWN   = "Error shutting down: "
	MAIN_ERROR_IS_DIR     = "Error provided path is a directory"

	HOME = "~"

	ERROR_DATABASE_NOT_FOUND = "Database is not initialized"
	ERROR_DATABASE_CLOSED    = "Database is closed"
	ERROR_KEY_NOT_FOUND      = "Key is not found: "

	// Restrouter
	ERROR_MODE              = "Backup Mode:"
	ERROR_GIT               = "GIT Mode:"
	ERROR_STATUS            = "GetStatus:"
	ERROR_ISSEALED          = "IsSealed:"
	ERROR_UNSEAL            = "Unseal:"
	ERROR_SEAL              = "Seal:"
	ERROR_RUNBACKUP         = "RunBackupJob:"
	ERROR_RUNMOUNT          = "RunMountJob:"
	ERROR_CONFIG            = "GetConfigFromVault:"
	ERROR_BINDING           = "BindJSON:"
	ERROR_PUT_TOKEN         = "PutToken:"
	ERROR_PUT_SEAL_KEY      = "PutSealKey:"
	REST_JSON_MESSAGE       = "message"
	REST_VAULT_SEAL_MESSAGE = "Vault seal is: "

	ERROR_VAULT_SEALED         = "Vault is sealed."
	ERROR_VAULT_NO_SECRET      = "Vault has no data for this endpoint."
	ERROR_VAULT_CONFIG_MISSING = "Vault config is missing"

	ERROR_UNMARSHAL        = "Error marshaling message: "
	ERROR_SENDING_REQUEST  = "Error sending request: "
	ERROR_READING_RESPONSE = "Error reading response: "
	ERROR_TIMESTAMP        = "Error retrieving timestamp: "

	BACKUP_TEST_FOLDER       = "~/test/Backup"
	BACKUP_TEST_EXCLUDE_FILE = "~/test/exclude\n~/*.go"
	BACKUP_TEST_CONF_FILE    = "~/test/Backup/config"

	// Mount Constants For Tests
	GOCRYPT_TEST_MOUNTPATH = "~/test/tmp-mount"
	GOCRYPT_TEST_FILE      = "/test"
	GOCRYPT_TEST_FOLDER    = "~/test/tmp"

	GIT_TEST_FOLDER       = "~/test/reverse"
	GIT_TEST_REPO         = "https://github.com/azak-azkaran/reverse-link"
	GIT_TEST_FOLDER_VIMRC = "~/test/vimrc"
	GIT_TEST_REPO_VIMRC   = "https://github.com/amix/vimrc.git"
	GIT_TEST_COMMIT       = "dd7e3c0b1ec7bbde6034d8cb2739bcd67f2530a4"

	VAULT_TEST_PASSWORD            = "hallo"
	VAULT_TEST_TOKEN               = "superrandompasswordtoken"
	VAULT_TEST_PATH                = "~/test/tmp"
	VAULT_TEST_MOUNTPATH           = "~/test/tmp-mount"
	VAULT_TEST_CONFIGPATH          = "gocryptpath"
	VAULT_TEST_BACKUP_PATH         = "~/test/Backup"
	VAULT_TEST_BACKUP_EXCLUDE_FILE = "~/test/exclude"
	VAULT_TEST_BACKUP_SECRET_KEY   = "secret.key"
	VAULT_TEST_BACKUP_ACCESS_KEY   = "access.key"
	VAULT_TEST_ROLE_ID             = "approleid"
	VAULT_TEST_SECRET_ID           = "appsecretid"

	MAIN_TEST_ADDRESS        = "localhost:8031"
	MAIN_TEST_PATHDB         = "./test/DB"
	MAIN_TEST_DURATION       = "1h30m"
	MAIN_TEST_MOUNT_DURATION = "5s"
	MAIN_TEST_MOUNT_ALLOW    = "false"
	MAIN_TEST_KEYFILE_PATH   = "./test/keyfile"

	REST_TEST_TOKEN      = "http://localhost:8031/token"
	REST_TEST_LOG        = "http://localhost:8031/logs"
	REST_TEST_BACKUP     = "http://localhost:8031/backup"
	REST_TEST_STATUS     = "http://localhost:8031/status"
	REST_TEST_MOUNT      = "http://localhost:8031/mount"
	REST_TEST_GIT        = "http://localhost:8031/git"
	REST_TEST_UNSEAL     = "http://localhost:8031/unseal"
	REST_TEST_IS_SEALED  = "http://localhost:8031/is_sealed"
	REST_TEST_SEAL       = "http://localhost:8031/seal"
	REST_TEST_PING       = "http://localhost:8031/ping"
	REST_TEST_UNSEAL_KEY = "http://localhost:8031/unsealkey"
)
