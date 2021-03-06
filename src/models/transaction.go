package models

import (
	"os"
	"strconv"

	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
)

const (
	DefaultTransactionAttempts = 5
)

func GetTransactionAttemptsFromEnvWithName(name string) int {
	v := os.Getenv(name)
	if v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return DefaultTransactionAttempts
		}
		return i
	}
	return DefaultTransactionAttempts
}

func GetTransactionAttemptsFromEnv() int {
	return GetTransactionAttemptsFromEnvWithName("DEFAULT_TRANSACTION_ATTEMPTS")
}

func GetTransactionOptions() *datastore.TransactionOptions {
	return &datastore.TransactionOptions{XG: false, Attempts: GetTransactionAttemptsFromEnv()}
}

func GetTransactionOptionsWithXG() *datastore.TransactionOptions {
	return &datastore.TransactionOptions{XG: true, Attempts: GetTransactionAttemptsFromEnv()}
}
