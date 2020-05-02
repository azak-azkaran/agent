package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type AddCryptFolder struct {
	cryptoFolder string `json:"cryptoFolder"`
	mountFolder  string `json:"mountFolder"`
	password     string `json:"password"`
}

func CreateRestHandler() http.Handler {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	return r
}
