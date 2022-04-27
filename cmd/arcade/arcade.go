package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	arcadehttp "github.com/homedepot/arcade/internal/http"
	"github.com/homedepot/arcade/internal/middleware"
)

var (
	r = gin.Default()
)

func init() {
	gin.ForceConsoleColor()

	var (
		controller arcadehttp.Controller
		err        error
	)

	if dir := os.Getenv("ARCADE_CONFIG_DIRECTORY"); dir != "" {
		controller, err = arcadehttp.NewController(dir)
	} else {
		controller, err = arcadehttp.NewDefaultController()
	}

	if err != nil {
		log.Fatal(err)
	}

	apiKey := mustGetenv("ARCADE_API_KEY")
	r.Use(middleware.NewAPIKeyAuth(apiKey))

	r.GET("/tokens", controller.GetToken)
}

func mustGetenv(env string) (s string) {
	if s = os.Getenv(env); s == "" {
		log.Fatal(env + " not set; exiting.")
	}

	return
}

// Run arcade on port 1982.
func main() {
	if err := r.Run(":1982"); err != nil {
		log.Fatal(err)
	}
}
