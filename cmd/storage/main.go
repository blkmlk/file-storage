package main

import (
	"log"
	"net"

	"github.com/blkmlk/file-storage/env"

	"github.com/blkmlk/file-storage/protocol"

	"google.golang.org/grpc"

	"github.com/blkmlk/file-storage/internal/services/filestorage"
	"github.com/blkmlk/file-storage/internal/services/storage"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	container.Provide(filestorage.NewFS)
	container.Provide(storage.New)

	var fStorage *storage.Storage
	err := container.Invoke(func(s *storage.Storage) {
		fStorage = s
	})
	if err != nil {
		log.Fatal(err)
	}

	host, err := env.Get(env.HOST)
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	protocol.RegisterStorageServer(server, fStorage)

	log.Printf("listening to %s...", host)
	if err = server.Serve(l); err != nil {
		log.Fatal(err)
	}
}
