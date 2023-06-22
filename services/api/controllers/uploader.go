package controllers

import "github.com/gin-gonic/gin"

type UploaderController struct {
}

func NewUploadController() (*UploaderController, error) {
	return &UploaderController{}, nil
}

func (c *UploaderController) UploadFile(ctx *gin.Context) {
}
