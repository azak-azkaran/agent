package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogInit(t *testing.T) {

	message := "message"

	Sugar.Info(message)
	Sugar.Error(message)

	r := gin.New()
	r.Use(gin.LoggerWithConfig(*logconfig))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	server := &http.Server{
		Addr:    "localhost:8031",
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		assert.EqualError(t, err, http.ErrServerClosed.Error())
	}()

	time.Sleep(10 * time.Millisecond)

	sendingGet(t, REST_TEST_PING, http.StatusOK)

	err := server.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestLogZapLogger(t *testing.T) {
	fmt.Println("Running: TestZapLogger")

	atom := zap.NewAtomicLevel()
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.Encoding = "console"
	config.Level = atom

	logger, _ := config.Build()

	url := "www.google.de"
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	atom.SetLevel(zap.WarnLevel)
	sugar.Info("Now logs should be colored")
	sugar.Warnw("failed to fetch URL",
		// Structured context as loosely typed key-value pairs.
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Errorf("Failed to fetch URL: %s", url)
}
