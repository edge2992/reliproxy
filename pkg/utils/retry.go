package utils

import (
	"fmt"
	"time"
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
		time.Sleep(backoffDuration)
	}
	return nil, fmt.Errorf("operation failed after %d retries: %v", maxRetries, err)
}
