package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"reliproxy/pkg/httpclient"
	"reliproxy/pkg/utils"
)

type Handler struct {
	client httpclient.HttpClient
}

func NewHandler(client httpclient.HttpClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) HandleRequest(c *gin.Context) {
	resp, err := h.client.Get("https://api.thirdparty.com/data")

	if err != nil {
		switch {
		case errors.Is(err, utils.ErrRateLimitExceeded):
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Warn("Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		case errors.Is(err, utils.ErrUnexpectedStatusCode):
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unexpected status code")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected status code"})
		default:
			utils.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Internal server error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}

		c.Error(err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to read response body")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}
	resp.Body.Close()

	utils.Logger.WithFields(logrus.Fields{
		"status": resp.StatusCode,
		"body":   string(bodyBytes),
	}).Info("Request handled successfully")
	c.JSON(http.StatusOK, gin.H{"data": string(bodyBytes)})
}
