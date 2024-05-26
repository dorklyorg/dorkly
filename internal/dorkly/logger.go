package dorkly

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	logger = newLogger()
)

func newLogger() *zap.SugaredLogger {
	fileLoggerConfig := zap.NewDevelopmentConfig()
	fileLoggerConfig.OutputPaths = []string{"dorkly.log"}
	fileLogger, err := fileLoggerConfig.Build()
	if err != nil {
		fmt.Println("failed to create logger: ", err)
		panic(err)
	}

	consoleLoggerEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "",
		LevelKey:       "",
		NameKey:        "",
		CallerKey:      "",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleLogger := zapcore.NewCore(zapcore.NewConsoleEncoder(consoleLoggerEncoderConfig), zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	tee := zapcore.NewTee(fileLogger.Core(), consoleLogger)
	return zap.New(tee).Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
