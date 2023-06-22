package main

import (
	"github.com/blkmlk/file-storage/deps"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	container.Provide(deps.NewDB)
}
