package main

import (
	"github.com/blkmlk/file-storage/deps"
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

	if err := listener.Start(); err != nil {
		panic(err)
	}
}
