package controllers

import (
	"net/http"

	"github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/blkmlk/file-storage/services/splitter"
	"github.com/gin-gonic/gin"
)

type RestController struct {
	repo     repository.Repository
	splitter splitter.Splitter
}

func NewUploadController(
	repo repository.Repository,
	splitter splitter.Splitter,
) (*RestController, error) {
	return &RestController{
		repo:     repo,
		splitter: splitter,
	}, nil
}

func (c *RestController) UploadFile(ctx *gin.Context) {
	mp, err := ctx.FormFile("file")
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	uploader, err := c.splitter.GetUploader(ctx, splitter.GetUploaderInput{
		Name:        mp.Filename,
		Size:        mp.Size,
		MinStorages: 5,
		NumStorages: 6,
	})
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	fp, err := mp.Open()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if err = uploader.Upload(ctx, fp); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
	return
}
