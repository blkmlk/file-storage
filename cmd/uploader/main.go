package main

import (
	"log"

	"github.com/blkmlk/file-storage/deps"
	"github.com/blkmlk/file-storage/env"
	"github.com/blkmlk/file-storage/internal/services/api"
	controllers2 "github.com/blkmlk/file-storage/internal/services/api/controllers"
	"github.com/blkmlk/file-storage/internal/services/manager"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	container.Provide(deps.NewDB)
	container.Provide(controllers2.NewUploadController)
	container.Provide(controllers2.NewProtocolController)
	container.Provide(api.New)
	container.Provide(manager.New)
	container.Provide(manager.NewGRPCClientFactory)

	var listener api.API
	container.Invoke(func(a api.API) {
		listener = a
	})

	restHost, err := env.Get(env.RestHost)
	if err != nil {
		log.Fatal(err)
	}

	protocolHost, err := env.Get(env.ProtocolHost)
	if err != nil {
		log.Fatal(err)
	}

	if err := listener.Start(restHost, protocolHost); err != nil {
		panic(err)
	}
}
