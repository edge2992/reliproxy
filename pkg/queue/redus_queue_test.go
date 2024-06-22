package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestRedisQueue_Enqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockRedisClient(ctrl)
	queue := NewRedisQueue(mockClient, "request_queue")

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
}

func TestRedisQueue_Dequeue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockRedisClient(ctrl)
	queue := NewRedisQueue(mockClient, "request_queue")

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

	// リクエストデータの型アサーションを行い、正しい形式であることを確認
	resultData, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, requestData["key"], resultData["key"])
}
