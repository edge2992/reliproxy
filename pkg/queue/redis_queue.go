package queue

import (
	"context"
	"encoding/json"
	"time"
)

type RedisQueue struct {
	client    RedisClient
	queueName string
}

func NewRedisQueue(client RedisClient, queueName string) *RedisQueue {
	return &RedisQueue{client: client, queueName: queueName}
}

type Request struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

func (q *RedisQueue) Enqueue(requestID string, requestData interface{}) error {
	ctx := context.Background()
	request := Request{
		ID:   requestID,
		Data: requestData,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, q.queueName, string(data)).Err()
}

func (q *RedisQueue) Dequeue() (string, interface{}, error) {
	ctx := context.Background()
	result, err := q.client.BRPop(ctx, 0*time.Second, q.queueName).Result()
	if err != nil {
		return "", nil, err
	}
	var request Request
	if err := json.Unmarshal([]byte(result[1]), &request); err != nil {
		return "", nil, err
	}
	return request.ID, request.Data, nil
}
