package handlers

import (
	"errors"
	"net/http"
	"reliproxy/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandleError(c *gin.Context, err error) {
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
}
