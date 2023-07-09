package controllers

import (
	"context"

	repository2 "github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/blkmlk/file-storage/protocol"
)

type ProtocolController struct {
	protocol.UnimplementedUploaderServer
	repo repository2.Repository
}

func NewProtocolController(repo repository2.Repository) *ProtocolController {
	return &ProtocolController{
		repo: repo,
	}
}

func (p *ProtocolController) Register(ctx context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	storage := repository2.NewStorage(request.StorageId, request.Host)
	if err := p.repo.CreateOrUpdateStorage(ctx, &storage); err != nil {
		//	log
		return nil, err
	}
	return &protocol.RegisterResponse{}, nil
}
