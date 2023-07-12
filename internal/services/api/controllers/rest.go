package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/blkmlk/file-storage/env"

	"github.com/blkmlk/file-storage/internal/services/manager"

	"github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/gin-gonic/gin"
)

type RestController struct {
	repo           repository.Repository
	log            *zap.SugaredLogger
	fileManager    manager.Manager
	uploadFileHost string
}

type GetUploadLinkResponse struct {
	UploadLink string `json:"upload_link"`
}

func NewUploadController(
	repo repository.Repository,
	log *zap.SugaredLogger,
	fileManager manager.Manager,
) (*RestController, error) {
	uploadFileHost, err := env.Get(env.UploadFileHost)
	if err != nil {
		return nil, err
	}

	return &RestController{
		log:            log,
		repo:           repo,
		fileManager:    fileManager,
		uploadFileHost: uploadFileHost,
	}, nil
}

func (c *RestController) GetUploadLink(ctx *gin.Context) {
	id, err := c.fileManager.Prepare(ctx)
	if err != nil {
		c.log.With("err", err).Error("failed to prepare an upload link")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	uploadUrl := fmt.Sprintf("%s/%s", c.uploadFileHost, id)

	ctx.JSON(http.StatusCreated, &GetUploadLinkResponse{
		UploadLink: uploadUrl,
	})
}

func (c *RestController) PostUploadFile(ctx *gin.Context) {
	mf, err := ctx.MultipartForm()
	if err != nil {
		c.log.With("err", err).Error("failed to open multiform")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	files, ok := mf.File["file"]
	if !ok {
		c.log.With("err", err).Error("no file is provided in request")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if len(files) != 1 {
		c.log.With("err", err).Error("provided more than 1 file")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	file := files[0]

	id := ctx.Param("id")

	pipe, err := file.Open()
	if err != nil {
		c.log.With("err", err).Error("failed to open file")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	fileInfo := manager.FileInfo{
		Name:        file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Size:        file.Size,
	}

	err = c.fileManager.Store(ctx, id, fileInfo, pipe)
	if err != nil {
		switch {
		case errors.Is(err, manager.ErrBusy):
			ctx.String(http.StatusForbidden, "file is being stored")
		case errors.Is(err, manager.ErrExists):
			ctx.String(http.StatusForbidden, "file is stored")
		default:
			c.log.With("err", err).Error("failed to store")
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
	ctx.Status(http.StatusCreated)
}

func (c *RestController) GetDownloadFile(ctx *gin.Context) {
	fileName := ctx.Param("name")

	file, err := c.repo.GetFileByName(ctx, fileName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ctx.String(http.StatusForbidden, "file not found")
			return
		}
		c.log.With("err", err).Error("failed to get file")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
	}

	reader, err := c.fileManager.Load(ctx, fileName)
	if err != nil {
		if errors.Is(err, manager.ErrNotFound) {
			ctx.String(http.StatusForbidden, "file not found")
			return
		}
		c.log.With("err", err).Error("failed to load file")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.DataFromReader(http.StatusOK, file.Size, file.ContentType, reader, extraHeaders)
}
