package wookie

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/service"
)

type WookieService struct {
	service.BaseService
	Repository WookieRepository
}

func NewService(env service.Env, repository WookieRepository) WookieService {
	return WookieService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
	}
}

// Given a provided 'wookie',
func (svc WookieService) Compare() error {
	return nil
}

func (svc WookieService) Create(ctx context.Context) error {
	return nil
}

func (svc WookieService) IsFresh(ctx context.Context, wookie BasicToken) bool {
	return false
}
