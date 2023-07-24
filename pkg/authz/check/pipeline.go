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

	"github.com/rs/zerolog/log"
)

type Pipeline struct {
	serviceSemaphore chan struct{}
	subtaskSemaphore chan struct{}
}

func NewPipeline(maxServiceConcurrency int, maxSubtaskConcurrency int) *Pipeline {
	return &Pipeline{
		serviceSemaphore: make(chan struct{}, maxServiceConcurrency),
		subtaskSemaphore: make(chan struct{}, maxSubtaskConcurrency),
	}
}

func (p *Pipeline) AcquireServiceLock() {
	p.serviceSemaphore <- struct{}{}
}

func (p *Pipeline) ReleaseServiceLock() {
	<-p.serviceSemaphore
}

func (p *Pipeline) AnyOf(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	return p.execTasks(ctx, tasks, func(r Result, isLastExpected bool) (*Result, bool) {
		// Short-circuit - pick this result if it's a match
		if r.Matched {
			return &r, true
		}
		// Last result AND it's not a match due to prev condition -> return not matched
		if isLastExpected {
			return &Result{
				Matched:      false,
				DecisionPath: r.DecisionPath,
				Err:          nil,
			}, true
		}
		// Not a match, keep looking
		return nil, false
	})
}

func (p *Pipeline) AllOf(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	return p.execTasks(ctx, tasks, func(r Result, isLastExpected bool) (*Result, bool) {
		// Short-circuit - return not matched if any sub-result is not matched
		if !r.Matched {
			return &r, true
		}
		// Last result AND it's a match due to prev condition -> return matched
		if isLastExpected {
			return &Result{
				Matched:      true,
				DecisionPath: r.DecisionPath,
				Err:          nil,
			}, true
		}
		// Keep looking
		return nil, false
	})
}

func (p *Pipeline) NoneOf(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result)) Result {
	return p.execTasks(ctx, tasks, func(r Result, isLastExpected bool) (*Result, bool) {
		// Short-circuit - return not matched
		if r.Matched {
			return &Result{
				Matched:      false,
				DecisionPath: r.DecisionPath,
				Err:          nil,
			}, true
		}
		// Last result AND it's not a match due to prev condition -> return matched
		if isLastExpected {
			return &Result{
				Matched:      true,
				DecisionPath: r.DecisionPath,
				Err:          nil,
			}, true
		}
		// Keep looking
		return nil, false
	})
}

func (p *Pipeline) execTasks(ctx context.Context, tasks []func(execCtx context.Context, resultC chan<- Result), checkResultFunc func(r Result, isLastExpected bool) (*Result, bool)) Result {
	childContext, childCtxCancelFunc := context.WithCancel(ctx)
	defer childCtxCancelFunc()
	childResultC := make(chan Result, len(tasks))

	// Exec each task
	for _, t := range tasks {
		task := t
		p.subtaskSemaphore <- struct{}{}
		log.Ctx(ctx).Debug().Msgf("creating goroutine!")
		go func() {
			defer func() {
				<-p.subtaskSemaphore
			}()
			task(childContext, childResultC)
		}()
	}

	// Monitor results, short-circuit ret as needed
	resultsReceived := 0
	for result := range childResultC {
		if result.Err != nil {
			return result
		}
		resultsReceived++
		r, returnResult := checkResultFunc(result, resultsReceived == len(tasks))
		if returnResult {
			return *r
		}
	}
	// TODO: should prob return an err here
	return Result{
		Matched:      false,
		DecisionPath: nil,
		Err:          nil,
	}
}
