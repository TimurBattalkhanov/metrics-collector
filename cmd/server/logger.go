package main

import "go.uber.org/zap"

func newLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	sugarLogger := logger.Sugar()
	zap.ReplaceGlobals(logger)
	return sugarLogger, nil
}
