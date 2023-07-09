package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/blkmlk/file-storage/env"

	"github.com/blkmlk/file-storage/internal/services/manager"

	"github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/gin-gonic/gin"
)

type RestController struct {
	repo           repository.Repository
	fileManager    manager.Manager
	uploadFileHost string
}

type GetUploadLinkResponse struct {
	UploadLink string `json:"upload_link"`
}

func NewUploadController(
	repo repository.Repository,
	fileManager manager.Manager,
) (*RestController, error) {
	uploadFileHost, err := env.Get(env.UploadFileHost)
	if err != nil {
		return nil, err
	}

	return &RestController{
		repo:           repo,
		fileManager:    fileManager,
		uploadFileHost: uploadFileHost,
	}, nil
}

func (c *RestController) GetUploadLink(ctx *gin.Context) {
	id, err := c.fileManager.Prepare(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	uploadUrl := fmt.Sprintf("https://%s/api/v1/upload/%s", c.uploadFileHost, id)

	ctx.JSON(http.StatusCreated, GetUploadLinkResponse{
		UploadLink: uploadUrl,
	})
}

func (c *RestController) PostUploadFile(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	id := ctx.Param("id")

	pipe, err := file.Open()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	fileInfo := manager.FileInfo{
		Name: file.Filename,
		Size: file.Size,
	}

	err = c.fileManager.Store(ctx, id, fileInfo, pipe)
	if err != nil {
		if errors.Is(err, manager.ErrExists) {
		}
		if errors.Is(err, manager.ErrNotFound) {
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusCreated)
}

func (c *RestController) GetDownloadFile(ctx *gin.Context) {
	var buffer bytes.Buffer
	fileName := ctx.Param("name")

	file, err := c.repo.GetFileByName(ctx, fileName)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.DataFromReader(http.StatusOK, file.Size, "application/zip", &buffer, nil)

	if err = c.fileManager.Load(ctx, fileName, &buffer); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}
