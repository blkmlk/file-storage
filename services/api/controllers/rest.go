package controllers

import "github.com/gin-gonic/gin"

type RestController struct {
}

func NewUploadController() (*RestController, error) {
	return &RestController{}, nil
}

func (c *RestController) UploadFile(ctx *gin.Context) {
}
