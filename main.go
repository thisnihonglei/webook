package main

import (
	"github.com/gin-gonic/gin"
	"webbook/internal/web"
)

func main() {
	hdl := &web.UserHandler{}
	server := gin.Default()
	hdl.RegisterRoutes(server)
	server.Run(":8080")
}
