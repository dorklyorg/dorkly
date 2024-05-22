package dorkly

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	logger = newLogger()
)

func newLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("failed to create logger: ", err)
		panic(err)
	}
	return logger.Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
