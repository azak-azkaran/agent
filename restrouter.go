package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func RunRestServer(address string) *http.Server {
	server := &http.Server{
		Addr:    "localhost:8081",
		Handler: CreateRestHandler(),
	}
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			log.Println("Agent server closed happily...")
		} else if err != nil {
			log.Println("Agent server closed horribly...\n", err)
		}
	}()
	time.Sleep(1 * time.Millisecond)
	log.Println("Agent rest server startet on: ", server.Addr)
	return server
}

func CreateRestHandler() http.Handler {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/unseal", func(c *gin.Context) {

	})

	r.GET("/is_sealed", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "",
		})
	})

	r.PUT("/add_token", func(c *gin.Context) {

	})

	r.GET("/status", func(c *gin.Context) {

	})
	return r
}
