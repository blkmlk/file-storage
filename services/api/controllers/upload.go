package controllers

import "github.com/gin-gonic/gin"

type UploadController struct {
}

func NewUploadController() (*UploadController, error) {
	return &UploadController{}, nil
}

func (c *UploadController) UploadFile(ctx *gin.Context) {
}
