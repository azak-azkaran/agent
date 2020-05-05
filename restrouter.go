package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenMessage struct {
	token string
}

func postSeal(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	AgentConfiguration.Token = msg.token
	err = Seal(AgentConfiguration.VaultConfig, AgentConfiguration.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	AgentConfiguration.Token = msg.token
	c.JSON(http.StatusOK, gin.H{
		"message": true,
	})
}

func postAddToken(c *gin.Context) {
	var msg TokenMessage
	err := c.BindJSON(msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	AgentConfiguration.Token = msg.token
	c.JSON(http.StatusOK, gin.H{
		"message": "",
	})
}

func getIsSealed(c *gin.Context) {
	bool, err := IsSealed(AgentConfiguration.VaultConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": bool,
	})
}

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

	r.POST("/seal", postSeal)
	r.GET("/is_sealed", getIsSealed)
	r.POST("/add_token", postAddToken)
	r.GET("/status", func(c *gin.Context) {

	})
	return r
}
