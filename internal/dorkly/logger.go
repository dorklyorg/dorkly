package dorkly

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	logger = newLogger()
)

func newLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopmentConfig().Build()
	//fileLoggerConfig := zap.NewDevelopmentConfig()
	//fileLoggerConfig.OutputPaths = []string{"dorkly.log"}
	//fileLogger, err := fileLoggerConfig.Build()
	if err != nil {
		fmt.Println("failed to create logger: ", err)
		panic(err)
	}
	return logger.Sugar()

	// This is the original code that was replaced by the above code.. It was an attempt to have cleaner display output + more detailed logs saved to a file.
	/*
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
	*/

}

func GetLogger() *zap.SugaredLogger {
	return logger
}
