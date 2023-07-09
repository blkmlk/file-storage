package api

import (
	"net"

	controllers2 "github.com/blkmlk/file-storage/internal/services/api/controllers"

	"github.com/blkmlk/file-storage/env"
	"github.com/blkmlk/file-storage/protocol"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type API interface {
	Start() error
	Stop() error
}

const (
	PathGetUploadFile  = "/api/v1/upload"
	PathPostUploadFile = "/api/v1/upload/:id"
)

type api struct {
	restHost           string
	protocolHost       string
	restController     *controllers2.RestController
	protocolController *controllers2.ProtocolController
	restServer         *gin.Engine
	grpcServer         *grpc.Server
}

func New(
	restController *controllers2.RestController,
	protocolController *controllers2.ProtocolController,
) (API, error) {
	restHost, err := env.Get(env.RestHost)
	if err != nil {
		return nil, err
	}

	protocolHost, err := env.Get(env.ProtocolHost)
	if err != nil {
		return nil, err
	}

	a := api{
		restHost:           restHost,
		protocolHost:       protocolHost,
		restController:     restController,
		protocolController: protocolController,
		restServer:         gin.Default(),
	}

	a.initRest()
	a.initGrpc()

	return &a, nil
}

func (a *api) initRest() {
	a.restServer.GET(PathGetUploadFile, a.restController.GetUploadLink)
	a.restServer.POST(PathPostUploadFile, a.restController.PostUploadFile)
}

func (a *api) initGrpc() {
	a.grpcServer = grpc.NewServer()
	protocol.RegisterUploaderServer(a.grpcServer, a.protocolController)
}

func (a *api) Start() error {
	listener, err := net.Listen("tcp", a.protocolHost)
	if err != nil {
		return err
	}

	errs := make(chan error)

	go func() {
		errs <- a.grpcServer.Serve(listener)
	}()

	go func() {
		errs <- a.restServer.Run(a.restHost)
	}()

	return <-errs
}

func (a *api) Stop() error {
	a.grpcServer.Stop()

	return nil
}
