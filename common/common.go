package common

const (
	CONTAINER_STOP_TIMEOUT = 15
	PRODUCTION             = "PROD"
	TESTING                = "TEST"

	DAEMON_TYPE  = "daemon"
	TEST_TYPE    = "test"
	DEFAULT_TYPE = "app"
	BUILD_TYPE   = "build"
	PULL_TYPE    = "pull"

	STATUS_IDENT = "__STATUS__"
	STATUS_DIE   = "die"
	STATUS_START = "start"

	ADD_TASK    = 1
	REMOVE_TASK = 2
	BUILD_TASK  = 3
	INFO_TASK   = 4
	TEST_TASK   = 5
	UPDATE_TASK = 6

	PROD_CONFIG_FILE = "resource-prod"
	TEST_CONFIG_FILE = "resource-test"
)
