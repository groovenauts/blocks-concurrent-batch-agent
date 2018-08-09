package models

import (
	"context"
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

func GetTransactionOptions(ctx context.Context) *datastore.TransactionOptions {
	opts := datastore.TransactionOptions{
		XG:       false,
		Attempts: GetTransactionAttemptsFromEnv(),
	}
	// log.Debugf(ctx, "TransactionOptions: %v\n", opts)
	return &opts
}
