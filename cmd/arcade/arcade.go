package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/internal/google"
	arcadehttp "github.com/homedepot/arcade/internal/http"
	"github.com/homedepot/arcade/internal/microsoft"
	"github.com/homedepot/arcade/internal/middleware"
	"github.com/homedepot/arcade/internal/rancher"
)

var (
	r = gin.Default()
)

func init() {
	controller := &arcadehttp.Controller{}

	gin.ForceConsoleColor()

	apiKey := mustGetenv("ARCADE_API_KEY")
	r.Use(middleware.NewApiKeyAuth(apiKey))

	controller.GoogleClient = google.NewClient()

	if s := os.Getenv("RANCHER_ENABLED"); s == "TRUE" {
		rancherClient := mustInstantiateRancherClient()
		controller.RancherClient = rancherClient
	}

	if s := os.Getenv("MICROSOFT_ENABLED"); s == "TRUE" {
		microsoftClient := mustInstantiateMicrosoftClient()
		controller.MicrosoftClient = microsoftClient
	}

	r.GET("/tokens", controller.GetToken)
}

func mustGetenv(env string) (s string) {
	if s = os.Getenv(env); s == "" {
		log.Fatal(env + " not set; exiting.")
	}

	return
}

func mustInstantiateRancherClient() *rancher.Client {
	url := mustGetenv("RANCHER_URL")
	username := mustGetenv("RANCHER_USERNAME")
	password := mustGetenv("RANCHER_PASSWORD")

	rancherClient := rancher.NewClient()
	rancherClient.WithURL(url)
	rancherClient.WithUsername(username)
	rancherClient.WithPassword(password)

	if caCerts := os.Getenv("RANCHER_CACERTS"); caCerts != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		rootCAs.AppendCertsFromPEM([]byte(caCerts))

		t := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		}

		rancherClient.WithTransport(t)
	}

	return rancherClient
}

func mustInstantiateMicrosoftClient() *microsoft.Client {
	loginEndpoint := mustGetenv("MICROSOFT_LOGIN_ENDPOINT")
	clientID := mustGetenv("MICROSOFT_CLIENT_ID")
	clientSecret := mustGetenv("MICROSOFT_CLIENT_SECRET")
	resource := mustGetenv("MICROSOFT_RESOURCE")

	microsoftClient := microsoft.NewClient()
	microsoftClient.WithLoginEndpoint(loginEndpoint)
	microsoftClient.WithClientID(clientID)
	microsoftClient.WithClientSecret(clientSecret)
	microsoftClient.WithResource(resource)

	return microsoftClient
}

// Run arcade on port 1982.
func main() {
	if err := r.Run(":1982"); err != nil {
		log.Fatal(err)
	}
}
