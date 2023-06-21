package wookie

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

	// If client didn't pass a wookie, run readFunc() with existing ctx and return latest wookie if present
	clientWookie, hasClientWookie := ctx.Value(ClientTokenKey{}).(Token)
	if !hasClientWookie {
		// TODO: Ideally the server should default to some trailing wookie value here. For now, default to 'unsafe' op to always use up-to-date db.
		unsafeCtx := context.WithValue(ctx, database.UnsafeOp{}, true)
		writerLatest, e := svc.Repository.GetLatest(unsafeCtx)
		var latestWookieToReturn *Token
		if e != nil {
			log.Ctx(ctx).Warn().Err(e).Msg("error getting writer latest wookie")
			latestWookieToReturn = nil
		}
		var wkCtx context.Context
		if writerLatest != nil {
			latestWookieToReturn = writerLatest.ToToken()
			wkCtx = context.WithValue(unsafeCtx, wookieQueryContextKey{}, latestWookieToReturn)
		} else {
			wkCtx = unsafeCtx
		}
		e = readFunc(wkCtx)
		if e != nil {
			return nil, e
		}
		return latestWookieToReturn, nil
	}

	// Otherwise, compare client wookie to a reader's latest wookie to see if we can use it
	var latestWookieToReturn *Token
	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		unsafe := false

		// First, get the reader's latest wookie
		readerLatest, err := svc.Repository.GetLatest(connCtx)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("error getting reader latest wookie")
			unsafe = true
		}

		// Compare reader wookie against client-provided wookie
		if readerLatest != nil {
			if readerLatest.GetID() < clientWookie.ID {
				// Reader is behind so op is unsafe
				unsafe = true
			} else {
				// Reader is up-to-date or ahead so is safe to use
				unsafe = false
				latestWookieToReturn = readerLatest.ToToken()
			}
		}

		wkCtx := context.WithValue(connCtx, database.UnsafeOp{}, unsafe)
		if unsafe {
			// Get writer's latest wookie
			writerLatest, err := svc.Repository.GetLatest(wkCtx)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("error getting writer latest wookie")
				latestWookieToReturn = nil
			}
			if writerLatest != nil {
				latestWookieToReturn = writerLatest.ToToken()
			}
		}

		// Execute read
		readCtx := context.WithValue(wkCtx, wookieQueryContextKey{}, latestWookieToReturn)
		err = readFunc(readCtx)
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	return latestWookieToReturn, nil
}
