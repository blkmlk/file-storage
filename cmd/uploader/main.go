package main

import (
	"go.uber.org/zap"

	"github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/blkmlk/file-storage/internal/services/cache"

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
	container.Provide(repository.New)
	container.Provide(controllers2.NewUploadController)
	container.Provide(controllers2.NewProtocolController)
	container.Provide(api.New)
	container.Provide(manager.New)
	container.Provide(manager.NewGRPCClientFactory)
	container.Provide(cache.NewMapCache)
	container.Provide(deps.NewZapLogger)

	var listener api.API
	var log *zap.SugaredLogger
	err := container.Invoke(func(a api.API, l *zap.SugaredLogger) {
		listener = a
		log = l
	})
	if err != nil {
		log.Fatal(err)
	}

	restHost, err := env.Get(env.RestHost)
	if err != nil {
		log.Fatal(err)
	}

	protocolHost, err := env.Get(env.ProtocolHost)
	if err != nil {
		log.Fatal(err)
	}

	if err = listener.Start(restHost, protocolHost); err != nil {
		log.Fatal(err)
	}
}
