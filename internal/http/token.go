package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetToken returns a new access token for a given provider.
func (ctl *Controller) GetToken(c *gin.Context) {
	provider := c.Query("provider")

	switch provider {
	case "google", "":
		ctl.getGoogleToken(c)
	case "microsoft":
		ctl.getMicrosoftToken(c)
	case "rancher":
		ctl.getRancherToken(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported token provider: %s", provider)})

		return
	}
}

func (ctl *Controller) getGoogleToken(c *gin.Context) {
	t, err := ctl.GoogleClient.Token(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": t})
}

func (ctl *Controller) getMicrosoftToken(c *gin.Context) {
	if ctl.MicrosoftClient == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token provider not configured: microsoft"})

		return
	}

	t, err := ctl.MicrosoftClient.Token(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": t})
}

func (ctl *Controller) getRancherToken(c *gin.Context) {
	if ctl.RancherClient == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token provider not configured: rancher"})

		return
	}

	t, err := ctl.RancherClient.Token(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": t})
}
