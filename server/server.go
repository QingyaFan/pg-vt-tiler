package server

import (
	"github.com/gin-gonic/gin"

	"net/http"

	"pg-vt-tiler/engine"
)

// GetTile get tile
func GetTile()  {
	r := gin.Default()
	r.Get("/ping", func (c *gin.Context)  {
		c.JSON(200, gin.H{
			"message": "pong"
		})
	})

	r.Run()
}
