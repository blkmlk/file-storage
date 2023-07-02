package controllers

import (
	"context"

	"github.com/blkmlk/file-storage/services/repository"

	"github.com/blkmlk/file-storage/protocol"
)

type ProtocolController struct {
	protocol.UnimplementedUploaderServer
	repo repository.Repository
}

func NewProtocolController(repo repository.Repository) *ProtocolController {
	return &ProtocolController{
		repo: repo,
	}
}

func (p *ProtocolController) Register(ctx context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	storage := repository.NewStorage(request.StorageId, request.Host)
	if err := p.repo.CreateOrUpdateStorage(ctx, &storage); err != nil {
		//	log
		return nil, err
	}
	return &protocol.RegisterResponse{}, nil
}
