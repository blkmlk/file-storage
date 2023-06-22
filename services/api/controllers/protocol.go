package controllers

import (
	"context"

	"github.com/blkmlk/file-storage/protocol"
)

type ProtocolController struct {
	protocol.UnimplementedUploaderServer
}

func (p *ProtocolController) Register(ctx context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	return &protocol.RegisterResponse{}, nil
}

func NewProtocolController() *ProtocolController {
	return &ProtocolController{}
}
