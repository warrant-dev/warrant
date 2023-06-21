package wookie

import (
	"context"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

const currentWookieVersion = 1

type updateWookieKey struct{}

type wookieQueryContextKey struct{}

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
	_, hasQueryWookie := ctx.Value(wookieQueryContextKey{}).(*Token)
	if hasQueryWookie {
		return nil, errors.New("invalid state: can't call WookieUpdate() within WookieSafeRead()")
	}

	updateWookie, hasUpdateWookie := ctx.Value(updateWookieKey{}).(*Token)
	// An update is already in progress so continue with that ctx
	if hasUpdateWookie {
		e := updateFunc(ctx)
		if e != nil {
			return nil, e
		}
		return updateWookie, nil
	}

	// Otherwise create a new tx and new update wookie
	var newWookie *Token
	e := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		newWookieId, err := svc.Repository.Create(txCtx, currentWookieVersion)
		if err != nil {
			return err
		}
		token, err := svc.Repository.GetById(txCtx, newWookieId)
		if err != nil {
			return err
		}
		newWookie = token.ToToken()
		wkCtx := context.WithValue(txCtx, updateWookieKey{}, newWookie)
		err = updateFunc(wkCtx)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, e
	}
	return newWookie, nil
}

func (svc WookieService) WookieSafeRead(ctx context.Context, readFunc func(wkCtx context.Context) error) (*Token, error) {
	// A read is already in progress so continue with that ctx
	queryWookie, hasQueryWookie := ctx.Value(wookieQueryContextKey{}).(*Token)
	if hasQueryWookie {
		e := readFunc(ctx)
		if e != nil {
			return nil, e
		}
		return queryWookie, nil
	}

	// If client didn't pass a wookie, run readFunc() without any checks
	clientWookie, hasClientWookie := ctx.Value(ClientTokenKey{}).(Token)
	if !hasClientWookie {
		// TODO: Ideally the server should default to some trailing wookie value here. For now, default to 'unsafe' op to always use up-to-date db.
		unsafeCtx := context.WithValue(ctx, database.UnsafeOp{}, true)
		latest, e := svc.Repository.GetLatest(unsafeCtx)
		if e != nil {
			return nil, e
		}
		latestWookie := latest.ToToken()
		wkCtx := context.WithValue(unsafeCtx, wookieQueryContextKey{}, latestWookie)
		e = readFunc(wkCtx)
		if e != nil {
			return nil, e
		}
		return latestWookie, nil
	}

	// Otherwise, compare client wookie to a reader to see if we can use it
	var latestWookie *Token
	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		readerLatest, err := svc.Repository.GetLatest(connCtx)
		if err != nil {
			return err
		}
		var wkCtx context.Context
		if readerLatest.GetID() < clientWookie.ID {
			wkCtx = context.WithValue(connCtx, database.UnsafeOp{}, true)
		} else {
			wkCtx = context.WithValue(connCtx, database.UnsafeOp{}, false)
		}
		latest, err := svc.Repository.GetLatest(wkCtx)
		if err != nil {
			return err
		}
		latestWookie = latest.ToToken()
		readCtx := context.WithValue(wkCtx, wookieQueryContextKey{}, latestWookie)
		err = readFunc(readCtx)
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	return latestWookie, nil
}
