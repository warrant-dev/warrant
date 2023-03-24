package service

import "github.com/warrant-dev/warrant/pkg/database"

type Env interface {
	DB() database.Database
	EventDB() database.Database
}

type Service interface {
	GetRoutes() []Route
}

type BaseService struct {
	env Env
}

func (svc BaseService) Env() Env {
	return svc.env
}

func NewBaseService(env Env) BaseService {
	return BaseService{
		env: env,
	}
}
