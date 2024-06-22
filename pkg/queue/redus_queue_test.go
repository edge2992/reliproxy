package queue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestRedisQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockRedisClient(ctrl)
	queue := NewRedisQueue(mockClient, "request_queue")

	t.Run("Enqueue", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			requestID := "test-request-id"
			requestData := map[string]string{"key": "value"}
			request := Request{
				ID:   requestID,
				Data: requestData,
			}
			data, _ := json.Marshal(request)

			mockClient.EXPECT().LPush(gomock.Any(), "request_queue", string(data)).Return(redis.NewIntCmd(context.Background()))

			err := queue.Enqueue(requestID, requestData)
			assert.NoError(t, err)
		})

		t.Run("SerializationError", func(t *testing.T) {
			requestID := "test-request-id"
			requestData := make(chan int) // シリアライズできないデータ型

			err := queue.Enqueue(requestID, requestData)
			assert.Error(t, err)
		})

		t.Run("RedisConnectionError", func(t *testing.T) {
			requestID := "test-request-id"
			requestData := map[string]string{"key": "value"}
			request := Request{
				ID:   requestID,
				Data: requestData,
			}
			data, _ := json.Marshal(request)

			intCmd := redis.NewIntCmd(context.Background())
			intCmd.SetErr(errors.New("redis connection error"))
			mockClient.EXPECT().LPush(gomock.Any(), "request_queue", string(data)).Return(intCmd)

			err := queue.Enqueue(requestID, requestData)
			assert.Error(t, err)
		})
	})

	t.Run("Dequeue", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			requestID := "test-request-id"
			requestData := map[string]string{"key": "value"}
			request := Request{
				ID:   requestID,
				Data: requestData,
			}
			data, _ := json.Marshal(request)
			expectedResult := []string{"request_queue", string(data)}

			mockClient.EXPECT().BRPop(gomock.Any(), 0*time.Second, "request_queue").Return(redis.NewStringSliceResult(expectedResult, nil))

			dequeuedRequestID, result, err := queue.Dequeue()
			assert.NoError(t, err)
			assert.Equal(t, requestID, dequeuedRequestID)

			resultData, ok := result.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, requestData["key"], resultData["key"])
		})

		t.Run("RedisConnectionError", func(t *testing.T) {
			mockClient.EXPECT().BRPop(gomock.Any(), 0*time.Second, "request_queue").Return(redis.NewStringSliceResult(nil, errors.New("redis connection error")))

			_, _, err := queue.Dequeue()
			assert.Error(t, err)
		})

		t.Run("DeserializationError", func(t *testing.T) {
			invalidData := "invalid data"
			expectedResult := []string{"request_queue", invalidData}

			mockClient.EXPECT().BRPop(gomock.Any(), 0*time.Second, "request_queue").Return(redis.NewStringSliceResult(expectedResult, nil))

			_, _, err := queue.Dequeue()
			assert.Error(t, err)
		})
	})
}
