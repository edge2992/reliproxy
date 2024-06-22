package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reliproxy/pkg/queue"
	"reliproxy/pkg/repository"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestAsyncWriteHandler_HandleRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedisClient := queue.NewMockRedisClient(ctrl)
	queue := queue.NewRedisQueue(mockRedisClient, "request_queue")
	mockRepo := repository.NewMockRequestStatusRepository(ctrl)

	handler := NewAsyncWriteHandler(queue, mockRepo)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/request", handler.HandleRequest)

	t.Run("successful request", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any()).Return(nil)
		mockRedisClient.EXPECT().LPush(gomock.Any(), "request_queue", gomock.Any()).Return(redis.NewIntCmd(context.Background()))

		reqBody := bytes.NewBufferString(`{"data": "test"}`)
		req, _ := http.NewRequest("POST", "/request", reqBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"request_id"`)
	})

	t.Run("repository save failure", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any()).Return(assert.AnError)

		reqBody := bytes.NewBufferString(`{"data": "test"}`)
		req, _ := http.NewRequest("POST", "/request", reqBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Failed to save request status to db"}`, w.Body.String())
	})

	t.Run("queue enqueue failure", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any()).Return(nil)
		intCmd := redis.NewIntCmd(context.Background())
		intCmd.SetErr(errors.New("redis connection error"))
		mockRedisClient.EXPECT().LPush(gomock.Any(), "request_queue", gomock.Any()).Return(intCmd)

		reqBody := bytes.NewBufferString(`{"data": "test"}`)
		req, _ := http.NewRequest("POST", "/request", reqBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Failed to enqueue request"}`, w.Body.String())
	})
}
