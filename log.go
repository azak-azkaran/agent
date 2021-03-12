package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
)

var (
	Atom      = zap.NewAtomicLevel()
	LogConfig = zap.NewProductionConfig()
	logger    *zap.Logger
	Sugar     *zap.SugaredLogger
)

var logconfig *gin.LoggerConfig = LogInit()

func LogInit() *gin.LoggerConfig {
	var err error
	LogConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	LogConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	LogConfig.Encoding = "console"
	LogConfig.Level = Atom
	logger, err = LogConfig.Build()
	if err != nil {
		fmt.Println("Error building logger:", err)
	}
	defer logger.Sync() // flushes buffer, if any
	Sugar = logger.Sugar()

	gin.SetMode(gin.ReleaseMode)
	return &gin.LoggerConfig{
		Formatter: DefaultLogFormatter,
		SkipPaths: []string{"/debug/vars"},
		Output:    GetLogger().Writer(),
	}
}

func StartProfiler(){
  profiler.Start(profiler.Config{
        ApplicationName: "agent",
        ServerAddress:   "http://localhost:4040",
    })
}

func GetLogger() *log.Logger {
	return zap.NewStdLog(logger)
}

func EnableError() {
	Atom.SetLevel(zap.ErrorLevel)
}

func EnableWarning() {
	Atom.SetLevel(zap.WarnLevel)
}

func EnableInfo() {
	Atom.SetLevel(zap.InfoLevel)
}

func EnableDebug() {
	Atom.SetLevel(zap.DebugLevel)
}

// defaultLogFormatter is the default log format function Logger middleware uses.
var DefaultLogFormatter = func(param gin.LogFormatterParams) string {
	var methodColor, resetColor string
	//if param.IsOutputColor() {
	methodColor = param.MethodColor()
	resetColor = param.ResetColor()
	//}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("GIN: %s %-7s %s| %13v | %15s | %s\n%s",
		methodColor, param.Method, resetColor,
		param.Latency,
		param.Request.URL.Host,
		param.Path,
		param.ErrorMessage,
	)
}
