package main

import (
	"github.com/blkmlk/file-storage/deps"
	"github.com/blkmlk/file-storage/services/api"
	"github.com/blkmlk/file-storage/services/api/controllers"
	"github.com/blkmlk/file-storage/services/splitter"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	container.Provide(deps.NewDB)
	container.Provide(controllers.NewUploadController)
	container.Provide(controllers.NewProtocolController)
	container.Provide(api.New)
	container.Provide(splitter.New)

	var listener api.API
	container.Invoke(func(a api.API) {
		listener = a
	})

	if err := listener.Start(); err != nil {
		panic(err)
	}
}
