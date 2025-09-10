package services

import (
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.SugaredLogger
}

func NewService(logger *zap.SugaredLogger) *Service {
	return &Service{
		logger: logger,
	}
}
