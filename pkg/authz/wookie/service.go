// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wookie

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type wookieQueryContextKey struct{}

type WookieService struct {
	service.BaseService
	Repository WookieRepository
	Enabled    bool
}

func NewService(env service.Env, repository WookieRepository, enabled bool) *WookieService {
	return &WookieService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		Enabled:     enabled,
	}
}

func (svc WookieService) WookieSafeRead(ctx context.Context, readFunc func(wkCtx context.Context) error) (*Token, error) {
	// If wookies are explicitly disabled, just run the readFunc() without any wookie validation
	if !svc.Enabled {
		return nil, readFunc(ctx)
	}

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
	e := svc.Env().DB().ReplicaSafeRead(ctx, func(connCtx context.Context) error {
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
