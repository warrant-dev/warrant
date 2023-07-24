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

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

type Semaphore struct {
	semaphore chan struct{}
}

func NewSema(maxConcurrency int) *Semaphore {
	return &Semaphore{
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

func (s *Semaphore) Acquire() {
	s.semaphore <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.semaphore
}

type Result struct {
	Matched      bool
	DecisionPath []warrant.WarrantSpec
	Err          error
}

func anyOfBlocking(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	childContext, childCtxCancelFunc := context.WithCancel(ctx)
	defer childCtxCancelFunc()
	childResultC := make(chan Result, len(tasks))

	for _, t := range tasks {
		task := t
		go func() {
			task(childContext, childResultC)
		}()
	}

	resultsReceived := 0
	for result := range childResultC {
		resultsReceived++
		if result.Err != nil {
			return result
		}
		if result.Matched {
			return result
		}
		if resultsReceived == len(tasks) {
			return Result{
				Matched:      false,
				DecisionPath: result.DecisionPath,
				Err:          nil,
			}
		}
	}
	// TODO: should prob return an err here
	return Result{
		Matched:      false,
		DecisionPath: nil,
		Err:          nil,
	}
}

func allOfBlocking(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	childContext, childCtxCancelFunc := context.WithCancel(ctx)
	defer childCtxCancelFunc()
	childResultC := make(chan Result)

	for _, t := range tasks {
		task := t
		go func() {
			task(childContext, childResultC)
		}()
	}

	resultsReceived := 0
	for result := range childResultC {
		resultsReceived++
		if result.Err != nil {
			return result
		}
		if !result.Matched {
			return result
		}
		if resultsReceived == len(tasks) {
			return Result{
				Matched:      true,
				DecisionPath: result.DecisionPath,
				Err:          nil,
			}
		}
	}
	// TODO: should prob return an err here
	return Result{
		Matched:      false,
		DecisionPath: nil,
		Err:          nil,
	}
}

func noneOfBlocking(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	childContext, childCtxCancelFunc := context.WithCancel(ctx)
	defer childCtxCancelFunc()
	childResultC := make(chan Result)

	for _, t := range tasks {
		task := t
		go func() {
			task(childContext, childResultC)
		}()
	}

	resultsReceived := 0
	for result := range childResultC {
		resultsReceived++
		if result.Err != nil {
			return result
		}
		if result.Matched {
			return Result{
				Matched:      false,
				DecisionPath: result.DecisionPath,
				Err:          nil,
			}
		}
		if resultsReceived == len(tasks) {
			return Result{
				Matched:      true,
				DecisionPath: result.DecisionPath,
				Err:          nil,
			}
		}
	}
	// TODO: should prob return an err here
	return Result{
		Matched:      false,
		DecisionPath: nil,
		Err:          nil,
	}
}
