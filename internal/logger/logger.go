// internal/logger/logger.go
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	L *zap.SugaredLogger
)

func Init(logFile string, level zapcore.Level) error {
	// 滚动文件配置
	lj := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB
		MaxBackups: 7,
		MaxAge:     30, // days
		Compress:   true,
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEnc := zapcore.NewConsoleEncoder(encCfg)
	fileEnc := zapcore.NewJSONEncoder(encCfg)

	consoleCore := zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), level)
	fileCore := zapcore.NewCore(fileEnc, zapcore.AddSync(lj), level)
	core := zapcore.NewTee(consoleCore, fileCore)

	baseLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	L = baseLogger.Sugar()
	zap.ReplaceGlobals(baseLogger)
	return nil
}

func Sync() { _ = L.Sync() }

func Infof(template string, args ...interface{}) {
	L.Infof(template, args...)
}

func Debugf(template string, args ...interface{}) {
	L.Debugf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	L.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	L.Fatalf(template, args...)
}
