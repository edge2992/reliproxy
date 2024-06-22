package handlers

import (
	"net/http"
	"reliproxy/pkg/queue"
	"reliproxy/pkg/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AsyncWriteHandler struct {
	queue            queue.Queue
	statusRepository repository.RequestStatusRepository
}

func NewAsyncWriteHandler(queue queue.Queue, repository repository.RequestStatusRepository) *AsyncWriteHandler {
	return &AsyncWriteHandler{
		queue:            queue,
		statusRepository: repository,
	}
}

func (h *AsyncWriteHandler) HandleRequest(c *gin.Context) {
	requestID := uuid.New().String()
	requestData := c.Request.Body

	requestStatus := repository.RequestStatus{
		ID:     requestID,
		Status: "queued",
	}
	err := h.statusRepository.Create(&requestStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save request status to db"})
		return
	}

	err = h.queue.Enqueue(requestID, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"request_id": requestID})
}
