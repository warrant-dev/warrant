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

// Apply given updateFunc() and create a new wookie for this update. Returns the new wookie token.
func (svc WookieService) WithWookieUpdate(ctx context.Context, updateFunc func(txCtx context.Context) error) (*Token, error) {
	var newWookie Token
	e := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		// First, apply given update within tx
		err := updateFunc(txCtx)
		if err != nil {
			return err
		}
		// Create new wookie in same tx
		newWookieId, err := svc.Repository.Create(txCtx, currentWookieVersion)
		if err != nil {
			return err
		}
		token, err := svc.Repository.GetById(txCtx, newWookieId)
		if err != nil {
			return err
		}
		newWookie = token.ToToken()
		return nil
	})
	if e != nil {
		return nil, e
	}
	return &newWookie, nil
}

func (svc WookieService) WookieSafeRead(ctx context.Context, readFunc func(wkCtx context.Context) error) (*Token, error) {
	return nil, nil
	// The wookie in this ctx has already been checked so rely on that value
	// if ctx.Value(database.UnsafeOp{}) != nil {
	// 	isUnsafe := ctx.Value(database.UnsafeOp{}).(bool)
	// 	err := readFunc(ctx)

	// }

	// ctx.Value(database.UnsafeOp{})
}

// TODO: this is making too many requests to check for wookie
func (svc WookieService) getWookieContext(ctx context.Context) (context.Context, *Token, error) {
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
