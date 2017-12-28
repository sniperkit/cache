package cacher

import (
	"errors"

	"github.com/sniperkit/logger"
	"github.com/sniperkit/logger/backends/zap"
)

var (
	logx              logger.Logger
	logxInitialFields logger.Fields
	logxConfig        logger.Config
)

func InitLogger(backend logger.Logger) (bool, error) {
	if logx == nil {
		logx = backend
		return true, nil
	}
	return false, errors.New("Logger was already initialized.")
}

func NewLogger() (logx logger.Logger, err error) {
	return logx, err
}

func init() {
	var err error

	logx, err = zap.New(
		&logger.Config{
			Backend:       "backend",
			Level:         "debug",
			Encoding:      "console",
			DisableCaller: false,
		})
	if err != nil {
		panic(err)
	}
}
