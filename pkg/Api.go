package pkg

import "go.uber.org/zap"

type Api struct {
	endpoint string
	log      *zap.Logger
}

func NewApi(logger *zap.Logger) *Api {
	return &Api{
		endpoint: "",
		log:      logger,
	}
}
