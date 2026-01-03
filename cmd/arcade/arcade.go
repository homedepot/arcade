package main

import (
	"context"
	"log"
	"time"

	"github.com/homedepot/arcade/internal/api"
	"github.com/homedepot/arcade/internal/api/providers"
	"github.com/sethvargo/go-envconfig"
)

type (
	Config struct {
		APIKey          string `env:"ARCADE_API_KEY, required"`
		ConfigDirectory string `env:"ARCADE_CONFIG_DIRECTORY"`
		Port            int    `env:"PORT, default=1982"`
	}
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	tokenizers, err := providers.LoadTokenizersFromDir(
		cfg.ConfigDirectory,
		time.Second*api.DefaultTimeoutSeconds,
	)
	if err != nil {
		log.Fatal(err)
	}

	apiHandler := api.NewHandler(
		api.WithAPIKey(cfg.APIKey),
		api.WithTokenizers(tokenizers),
	)

	server := api.NewServer(api.WithPort(cfg.Port), api.WithHandler(apiHandler))

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
