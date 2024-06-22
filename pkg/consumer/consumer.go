package consumer

import (
	"reliproxy/pkg/httpclient"
	"reliproxy/pkg/queue"
	"reliproxy/pkg/repository"
	"reliproxy/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type Consumer struct {
	queue            queue.Queue
	statusRepository repository.RequestStatusRepository
	client           httpclient.HttpClient
}

func NewConsumer(queue queue.Queue, repository repository.RequestStatusRepository, client httpclient.HttpClient) *Consumer {
	return &Consumer{
		queue:            queue,
		statusRepository: repository,
		client:           client,
	}
}

func (c *Consumer) Start() {
	for {
		requestID, requestData, err := c.queue.Dequeue()
		if err != nil {
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to dequeue request")
			time.Sleep(1 * time.Second)
			continue
		}

		go c.consume(requestID, requestData)

	}
}

func (c *Consumer) consume(requestID string, requestData interface{}) {
	_, err := c.client.Get("https://api.thirdparty.com/data")
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to make request")
		return
	}

	err = c.statusRepository.Create(&repository.RequestStatus{
		ID: requestID, Status: "processed"})
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to update request status")
	}
}
