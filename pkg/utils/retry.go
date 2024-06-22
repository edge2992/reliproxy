package utils

import (
	"time"

	"github.com/sirupsen/logrus"
)

// RetryWithExponentialBackoff retries the operation with exponential backoff
func RetryWithExponentialBackoff(operation func() (interface{}, error), maxRetries int) (interface{}, error) {
	var result interface{}
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		backoffDuration := time.Duration((1 << i)) * time.Second
		Logger.WithFields(logrus.Fields{
			"retry":      i + 1,
			"maxRetries": maxRetries,
			"error":      err,
		}).Warningf("Retry %d/%d failed. Retrying in %v", i+1, maxRetries, backoffDuration)
		time.Sleep(backoffDuration)
	}

	Logger.WithFields(logrus.Fields{
		"maxRetries": maxRetries,
		"error":      err,
	}).Errorf("operation failed after %d retries: %v", maxRetries, err)

	return nil, err
}
