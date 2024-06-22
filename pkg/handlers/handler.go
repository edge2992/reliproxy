package handlers

import (
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
		HandleError(c, err)
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
