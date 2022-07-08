package tronApi

import "go.uber.org/zap"

type Api struct {
	endpoint string
	log      *zap.Logger
}

func NewApi(nodeUrl string, logger *zap.Logger) *Api {
	return &Api{
		endpoint: nodeUrl,
		log:      logger,
	}
}
