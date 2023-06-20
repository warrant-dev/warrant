package wookie

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

const currentWookieVersion = 1

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

func (svc WookieService) Create(ctx context.Context) (*Token, error) {
	newWookieId, err := svc.Repository.Create(ctx, currentWookieVersion)
	if err != nil {
		return nil, err
	}

	newWookie, err := svc.Repository.GetById(ctx, newWookieId)
	if err != nil {
		return nil, err
	}

	token := newWookie.ToToken()
	return &token, nil
}

// TODO: this is making too many requests to check for wookie
func (svc WookieService) GetWookieContext(ctx context.Context) (context.Context, *Token, error) {
	latest, err := svc.Repository.GetLatest(ctx)
	if err != nil {
		return ctx, nil, err
	}
	latestToken := latest.ToToken()
	clientWookie, ok := ctx.Value(TokenKey{}).(Token)
	// TODO: If no/invalid client wookie, use some smart, up-to-date value. But for now, be strict and default to writer.
	if !ok {
		return context.WithValue(ctx, database.UnsafeOp{}, true), &latestToken, nil
	}
	client, err := svc.Repository.GetById(ctx, clientWookie.ID)
	if err != nil {
		return context.WithValue(ctx, database.UnsafeOp{}, true), &latestToken, nil
	}

	// If server not up-to-date, unsafe for read ops
	if latest.GetID() < client.GetID() {
		return context.WithValue(ctx, database.UnsafeOp{}, true), &latestToken, nil
	}
	return ctx, &latestToken, nil
}
