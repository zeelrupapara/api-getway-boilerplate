// By Emran A. Hamdan, Lead Architect
package logger

import (
	"fmt"
	"greenlync-api-gateway/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

//var localCfg *config.Config

// Uber logger is the best, if we find a faster so wrap it !
type Logger struct {
	Logger *zap.SugaredLogger
}

// NewProductionZapLogger will return a new production logger backed by zap
func NewLogger(cfg *config.Config) (*Logger, error) {
	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	conf.DisableCaller = true
	conf.DisableStacktrace = true
	fmt.Println(cfg.Logger.LogFile)

	//zapLogger, err := conf.Build(zap.WrapCore(zapCore))
	zapLogger, err := conf.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Logger.LogFile,
			MaxSize:    1, // megabytes
			MaxBackups: 30,
			MaxAge:     30, // days
		})

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zap.DebugLevel,
		)
		cores := zapcore.NewTee(c, core)

		return cores

	}))

	return &Logger{
		Logger: zapLogger.Sugar(),
	}, err
}

// Emran replace with closuer
// func zapCore(c zapcore.Core) zapcore.Core {

// 	// lumberjack.Logger is already safe for concurrent use, so we don't need to
// 	// lock it.
// 	// we always have the rotate file as our micro-service name
// 	w := zapcore.AddSync(&lumberjack.Logger{
// 		Filename:   "./logs/greenlync-api-gateway.log",
// 		MaxSize:    1, // megabytes
// 		MaxBackups: 30,
// 		MaxAge:     30, // days
// 	})

// 	core := zapcore.NewCore(
// 		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
// 		w,
// 		zap.DebugLevel,
// 	)
// 	cores := zapcore.NewTee(c, core)

// 	return cores
// }
