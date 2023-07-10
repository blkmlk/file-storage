package api

import (
	"net"

	controllers2 "github.com/blkmlk/file-storage/internal/services/api/controllers"

	"github.com/blkmlk/file-storage/protocol"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type API interface {
	Start(restHost, protocolHost string) error
	Stop() error
}

const (
	PathGetUploadFile   = "/api/v1/upload"
	PathPostUploadFile  = "/api/v1/upload/:id"
	PathGetDownloadFile = "/api/v1/download/:name"
)

type api struct {
	restController     *controllers2.RestController
	protocolController *controllers2.ProtocolController
	restServer         *gin.Engine
	grpcServer         *grpc.Server
}

func New(
	restController *controllers2.RestController,
	protocolController *controllers2.ProtocolController,
) (API, error) {

	a := api{
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
	a.restServer.GET(PathGetDownloadFile, a.restController.GetDownloadFile)
}

func (a *api) initGrpc() {
	a.grpcServer = grpc.NewServer()
	protocol.RegisterUploaderServer(a.grpcServer, a.protocolController)
}

func (a *api) Start(restHost, protocolHost string) error {
	listener, err := net.Listen("tcp", protocolHost)
	if err != nil {
		return err
	}

	errs := make(chan error)

	go func() {
		errs <- a.grpcServer.Serve(listener)
	}()

	go func() {
		errs <- a.restServer.Run(restHost)
	}()

	return <-errs
}

func (a *api) Stop() error {
	a.grpcServer.Stop()

	return nil
}
