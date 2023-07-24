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
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
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

// TODO: fix decisionPath
func check(level int, sema *Semaphore, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- Result, svc *CheckService, typesMap *objecttype.ObjectTypeMap) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled check[%d] [%s]", level, checkSpec)
		return
	default:
		log.Ctx(ctx).Debug().Msgf("exec check[%d] [%s]", level, checkSpec)
		// 1. Check for direct warrant match
		matchedWarrant, err := svc.getWithPolicyMatch(ctx, sema, checkSpec.CheckWarrantSpec)
		if err != nil {
			// log.Ctx(ctx).Err(err).Msgf("ERR CheckTask getWithPolicyMatch")
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          err,
			}
			return
		}
		if matchedWarrant != nil {
			resultC <- Result{
				Matched:      true,
				DecisionPath: append(currentPath, *matchedWarrant),
				Err:          nil,
			}
			return
		}

		// 2. Check through indirect/group warrants
		var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
		additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
			checkGroup(level+1, sema, execCtx, checkSpec, currentPath, resultC, svc, typesMap)
		})

		// 3. And/or defined rules for target relation
		objectTypeSpec, err := typesMap.GetByTypeId(checkSpec.ObjectType)
		if err != nil {
			//log.Ctx(ctx).Err(err).Msgf("ERR CheckTask GetByTypeId")
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          err,
			}
			return
		}
		if relationRule, ok := objectTypeSpec.Relations[checkSpec.Relation]; ok {
			additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
				checkRule(level+1, sema, execCtx, checkSpec, currentPath, resultC, svc, typesMap, &relationRule)
			})
		}

		resultC <- anyOfBlocking(ctx, additionalTasks)
	}
}

func checkGroup(level int, sema *Semaphore, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- Result, svc *CheckService, typesMap *objecttype.ObjectTypeMap) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled checkGroup[%d] [%s]", level, checkSpec)
		return
	default:
		log.Ctx(ctx).Debug().Msgf("exec checkGroup[%d] [%s]", level, checkSpec)
		warrants, err := svc.getMatchingSubjects(ctx, sema, typesMap, checkSpec.ObjectType, checkSpec.ObjectId, checkSpec.Relation, checkSpec.Context)
		if err != nil {
			//log.Ctx(ctx).Err(err).Msgf("ERR GroupCheckTask getMatchingSubjects")
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          err,
			}
			return
		}

		var matchingWarrants []warrant.WarrantSpec
		for _, w := range warrants {
			if w.Subject.Relation == "" {
				continue
			}
			matchingWarrants = append(matchingWarrants, w)
		}
		if len(matchingWarrants) == 0 {
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
		}
		var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
		for _, w := range matchingWarrants {
			matchingWarrant := w
			additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
				check(level+1, sema, execCtx, CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: matchingWarrant.Subject.ObjectType,
						ObjectId:   matchingWarrant.Subject.ObjectId,
						Relation:   matchingWarrant.Subject.Relation,
						Subject:    checkSpec.Subject,
						Context:    checkSpec.Context,
					},
					Debug: checkSpec.Debug,
				}, append(currentPath, matchingWarrant), resultC, svc, typesMap)
			})
		}
		resultC <- anyOfBlocking(ctx, additionalTasks)
	}
}

func checkRule(level int, sema *Semaphore, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- Result, svc *CheckService, typesMap *objecttype.ObjectTypeMap, rule *objecttype.RelationRule) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled checkRule[%d] [%s] [%s]", level, checkSpec, rule)
		return
	default:
		log.Ctx(ctx).Debug().Msgf("exec checkRule[%d] [%s] [%s]", level, checkSpec, rule)
		warrantSpec := checkSpec.CheckWarrantSpec
		if rule == nil {
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
			return
		}
		switch rule.InheritIf {
		case "":
			// No match found
			resultC <- Result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
		case objecttype.InheritIfAllOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
					checkRule(level+1, sema, execCtx, checkSpec, currentPath, resultC, svc, typesMap, &subRule)
				})
			}
			resultC <- allOfBlocking(ctx, additionalTasks)
		case objecttype.InheritIfAnyOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
					checkRule(level+1, sema, execCtx, checkSpec, currentPath, resultC, svc, typesMap, &subRule)
				})
			}
			resultC <- anyOfBlocking(ctx, additionalTasks)
		case objecttype.InheritIfNoneOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
					checkRule(level+1, sema, execCtx, checkSpec, currentPath, resultC, svc, typesMap, &subRule)
				})
			}
			resultC <- noneOfBlocking(ctx, additionalTasks)
		default:
			if rule.OfType == "" && rule.WithRelation == "" {
				check(level+1, sema, ctx, CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: warrantSpec.ObjectType,
						ObjectId:   warrantSpec.ObjectId,
						Relation:   rule.InheritIf,
						Subject:    warrantSpec.Subject,
						Context:    warrantSpec.Context,
					},
					Debug: checkSpec.Debug,
				}, currentPath, resultC, svc, typesMap)
				return
			}

			matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, sema, typesMap, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
			if err != nil {
				//log.Ctx(ctx).Err(err).Msgf("ERR CheckRuleTask getMatchingSubjectsBySubjectType")
				resultC <- Result{
					Matched:      false,
					DecisionPath: currentPath,
					Err:          err,
				}
				return
			}
			if len(matchingWarrants) == 0 {
				resultC <- Result{
					Matched:      false,
					DecisionPath: currentPath,
					Err:          nil,
				}
			}
			var additionalTasks []func(execCtx context.Context, resultC chan<- Result)
			for _, w := range matchingWarrants {
				matchingWarrant := w
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- Result) {
					check(level+1, sema, execCtx, CheckSpec{
						CheckWarrantSpec: CheckWarrantSpec{
							ObjectType: matchingWarrant.Subject.ObjectType,
							ObjectId:   matchingWarrant.Subject.ObjectId,
							Relation:   rule.InheritIf,
							Subject:    warrantSpec.Subject,
							Context:    warrantSpec.Context,
						},
						Debug: checkSpec.Debug,
					}, append(currentPath, matchingWarrant), resultC, svc, typesMap)
				})
			}
			resultC <- anyOfBlocking(ctx, additionalTasks)
		}
	}
}
