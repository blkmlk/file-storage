package api

import (
	"github.com/blkmlk/file-storage/env"
	"github.com/blkmlk/file-storage/services/api/controllers"
	"github.com/gin-gonic/gin"
)

type API interface {
	Start() error
	Stop() error
}

const (
	PathUploadFile = "/api/v1/upload"
)

type api struct {
	uploadHost       string
	uploadController *controllers.UploadController
	engine           *gin.Engine
}

func New(uploadController *controllers.UploadController) (API, error) {
	uploadHost, err := env.Get(env.UploadHost)
	if err != nil {
		return nil, err
	}

	a := api{
		uploadHost:       uploadHost,
		uploadController: uploadController,
		engine:           gin.Default(),
	}

	a.init()

	return a, nil
}

func (a api) init() {
	a.engine.POST(PathUploadFile, a.uploadController.UploadFile)
}

func (a api) Start() error {
	errs := make(chan error)
	go func() {
		errs <- a.engine.Run(a.uploadHost)
	}()

	return <-errs
}

func (a api) Stop() error {
	//TODO implement me
	panic("implement me")
}
