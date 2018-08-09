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

func GetTransactionAttemptsFromEnv() int {
	v := os.Getenv("DEFAULT_TRANSACTION_ATTEMPTS")
	if v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return DefaultTransactionAttempts
		}
		return i
	}
	return DefaultTransactionAttempts
}

func GetTransactionOptions() *datastore.TransactionOptions {
	return &datastore.TransactionOptions{XG: false, Attempts: GetTransactionAttemptsFromEnv()}
}
