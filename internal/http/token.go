package http

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/pkg/provider"
)

const (
	lifecycleWithDash = 3 // "-XX" = dash + 2-char lifecycle code
)

// GetToken returns a new access token for a given provider.
func (ctl *Controller) GetToken(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		providerName = "google"
	}

	if len(providerName) >= (len(ProviderTypeVaultK8s) + lifecycleWithDash) {
		if providerName[0:len(ProviderTypeVaultK8s)] == ProviderTypeVaultK8s {
			tokenizer, ok := ctl.Tokenizers[providerName[0:len(ProviderTypeVaultK8s)+lifecycleWithDash]]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported token provider: %s", providerName)})
				return
			}

			ctx := context.WithValue(c.Request.Context(), provider.ProviderKey, providerName)
			t, err := tokenizer.Token(ctx)

			if err != nil {
				log.Printf("arcade: GetToken: failed to get token from provider %s: %s\n", providerName, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

				return
			}

			c.JSON(http.StatusOK, gin.H{"token": t})

			return
		}
	}

	tokenizer, ok := ctl.Tokenizers[providerName]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported token provider: %s", providerName)})

		return
	}

	t, err := tokenizer.Token(context.Background())
	if err != nil {
		log.Printf("arcade: GetToken: failed to get token from provider %s: %s\n", providerName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": t})
}
