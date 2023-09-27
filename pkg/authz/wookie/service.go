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

package authz

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

const currentWookieVersion = 1

type WookieService struct {
	service.BaseService
	Repository WookieRepository
}

func NewService(env service.Env, repository WookieRepository) *WookieService {
	return &WookieService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
	}
}

func (svc WookieService) CreateNewWookie(ctx context.Context) (*wookie.Token, error) {
	var newWookie *wookie.Token

	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		newWookieId, err := svc.Repository.Create(txCtx, currentWookieVersion)
		if err != nil {
			return err
		}
		wookie, err := svc.Repository.GetById(txCtx, newWookieId)
		if err != nil {
			return err
		}
		newWookie = wookie.ToToken()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newWookie, nil
}

func (svc WookieService) GetById(ctx context.Context, id int64) (*wookie.Token, error) {
	wookie, err := svc.Repository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	wookieToken := wookie.ToToken()

	return wookieToken, nil
}

func (svc WookieService) GetLatestWookie(ctx context.Context) (*wookie.Token, error) {
	latestWookie, err := svc.Repository.GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	latestWookieToken := latestWookie.ToToken()

	return latestWookieToken, nil
}

func (svc WookieService) WithNewWookie(ctx context.Context, txWookieFunc func(txCtx context.Context) error) (*wookie.Token, error) {
	serverCreatedWookie, hasServerCreatedWookie := ctx.Value(wookie.ServerCreatedWookieCtxKey{}).(*wookie.Token)
	// An update is already in progress so continue with that ctx
	if hasServerCreatedWookie {
		e := txWookieFunc(ctx)
		if e != nil {
			return nil, e
		}
		return serverCreatedWookie, nil
	}

	// Otherwise, create a new tx and a new wookie for writes in txWookieFunc to use.
	var newWookie *wookie.Token
	var err error
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		newWookie, err = svc.CreateNewWookie(txCtx)
		if err != nil {
			return err
		}

		wkCtx := context.WithValue(txCtx, wookie.ServerCreatedWookieCtxKey{}, newWookie)
		err = txWookieFunc(wkCtx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newWookie, nil
}
