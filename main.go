package main

import (
	"fmt"
	"os"

	"github.com/ahmad-khatib0-org/megacommerce-inventory/internal/server"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	if env != "dev" && env != "test" && env != "prod" {
		env = "dev"
	}

	config, err := server.LoadServiceConfig(fmt.Sprintf("config.%s.yaml", env))
	if err != nil {
		panic(err)
	}

	logger, err := logger.InitLogger(config.Service.Env)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	srv := &server.ServerArgs{Log: logger, Cfg: config}
	server.RunServer(srv)
}
