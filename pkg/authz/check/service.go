// Copyright 2024 WorkOS, Inc.
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
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/stats"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

const (
	MaxWarrants = 10000
)

type CheckContextFunc func(ctx context.Context) (context.Context, error)

type CheckService struct {
	service.BaseService
	warrantSvc         warrant.Service
	objectTypeSvc      objecttype.Service
	checkConfig        *config.CheckConfig
	createCheckContext CheckContextFunc
}

func defaultCreateCheckContext(ctx context.Context) (context.Context, error) {
	checkCtx := stats.BlankContextWithRequestStats(ctx)
	if wookie.ContainsLatest(ctx) {
		return wookie.WithLatest(checkCtx), nil
	}
	return checkCtx, nil
}

func NewService(env service.Env, warrantSvc warrant.Service, objectTypeSvc objecttype.Service, checkConfig *config.CheckConfig, checkContext CheckContextFunc) *CheckService {
	svc := &CheckService{
		BaseService:        service.NewBaseService(env),
		warrantSvc:         warrantSvc,
		objectTypeSvc:      objectTypeSvc,
		checkConfig:        checkConfig,
		createCheckContext: checkContext,
	}

	if checkContext == nil {
		svc.createCheckContext = defaultCreateCheckContext
	}

	return svc
}

func (svc CheckService) getWithPolicyMatch(ctx context.Context, checkPipeline *pipeline, spec CheckWarrantSpec) (*warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	listParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	listParams.WithLimit(MaxWarrants)
	warrantSpecs, _, _, err := svc.warrantSvc.List(
		ctx,
		warrant.FilterParams{
			ObjectType:      spec.ObjectType,
			ObjectId:        spec.ObjectId,
			Relation:        spec.Relation,
			SubjectType:     spec.Subject.ObjectType,
			SubjectId:       spec.Subject.ObjectId,
			SubjectRelation: spec.Subject.Relation,
		},
		listParams,
	)
	if err != nil || len(warrantSpecs) == 0 {
		return nil, err
	}

	// if a warrant without a policy is found, match it
	for _, w := range warrantSpecs {
		if w.Policy == "" {
			return &w, nil
		}
	}

	for _, w := range warrantSpecs {
		if w.Policy != "" {
			if policyMatched := evalWarrantPolicy(w, spec.Context); policyMatched {
				return &w, nil
			}
		}
	}

	return nil, nil
}

func (svc CheckService) getMatchingSubjects(ctx context.Context, checkPipeline *pipeline, objectType string, objectId string, relation string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	listParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	listParams.WithLimit(MaxWarrants)
	warrantSpecs, _, _, err = svc.warrantSvc.List(
		ctx,
		warrant.FilterParams{
			ObjectType: objectType,
			ObjectId:   objectId,
			Relation:   relation,
		},
		listParams,
	)
	if err != nil {
		return warrantSpecs, err
	}

	matchingSpecs := make([]warrant.WarrantSpec, 0)
	for _, w := range warrantSpecs {
		if w.Policy == "" {
			matchingSpecs = append(matchingSpecs, w)
		} else {
			if policyMatched := evalWarrantPolicy(w, checkCtx); policyMatched {
				matchingSpecs = append(matchingSpecs, w)
			}
		}
	}

	return matchingSpecs, nil
}

func (svc CheckService) getMatchingSubjectsBySubjectType(ctx context.Context, checkPipeline *pipeline, objectType string,
	objectId string, relation string, subjectType string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := svc.objectTypeSvc.GetByTypeId(ctx, objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	listParams := service.DefaultListParams(warrant.WarrantListParamParser{})
	listParams.WithLimit(MaxWarrants)
	warrantSpecs, _, _, err = svc.warrantSvc.List(
		ctx,
		warrant.FilterParams{
			ObjectType:  objectType,
			ObjectId:    objectId,
			Relation:    relation,
			SubjectType: subjectType,
		},
		listParams,
	)
	if err != nil {
		return warrantSpecs, err
	}

	matchingSpecs := make([]warrant.WarrantSpec, 0)
	for _, w := range warrantSpecs {
		if w.Policy == "" {
			matchingSpecs = append(matchingSpecs, w)
		} else {
			if policyMatched := evalWarrantPolicy(w, checkCtx); policyMatched {
				matchingSpecs = append(matchingSpecs, w)
			}
		}
	}

	return matchingSpecs, nil
}

func (svc CheckService) CheckMany(ctx context.Context, authInfo *service.AuthInfo, warrantCheck *CheckManySpec) (*CheckResultSpec, error) {
	start := time.Now().UTC()
	if warrantCheck.Op != "" && warrantCheck.Op != objecttype.InheritIfAllOf && warrantCheck.Op != objecttype.InheritIfAnyOf {
		return nil, service.NewInvalidParameterError("op", "must be either anyOf or allOf")
	}

	var checkResult CheckResultSpec
	checkResult.DecisionPath = make(map[string][]warrant.WarrantSpec, 0)

	if warrantCheck.Op == objecttype.InheritIfAllOf {
		var processingTime int64
		var isImplicit bool
		for _, warrantSpec := range warrantCheck.Warrants {
			match, decisionPath, implicit, err := svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: warrantSpec,
				Debug:            warrantCheck.Debug,
			})
			if err != nil {
				return nil, err
			}

			isImplicit = isImplicit || implicit
			if warrantCheck.Debug {
				checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
				if len(decisionPath) > 0 {
					checkResult.DecisionPath[warrantSpec.String()] = decisionPath
				}
			}

			if !match {
				checkResult.Code = http.StatusForbidden
				checkResult.Result = NotAuthorized
				checkResult.IsImplicit = false
				return &checkResult, nil
			}
		}

		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		checkResult.IsImplicit = isImplicit
		return &checkResult, nil
	}

	if warrantCheck.Op == objecttype.InheritIfAnyOf {
		var processingTime int64
		for _, warrantSpec := range warrantCheck.Warrants {
			match, decisionPath, isImplicit, err := svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: warrantSpec,
				Debug:            warrantCheck.Debug,
			})
			if err != nil {
				return nil, err
			}

			if warrantCheck.Debug {
				checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
				if len(decisionPath) > 0 {
					checkResult.DecisionPath[warrantSpec.String()] = decisionPath
				}
			}

			if match {
				checkResult.Code = http.StatusOK
				checkResult.Result = Authorized
				checkResult.IsImplicit = isImplicit
				return &checkResult, nil
			}
		}

		checkResult.Code = http.StatusForbidden
		checkResult.Result = NotAuthorized
		checkResult.IsImplicit = false
		return &checkResult, nil
	}

	if len(warrantCheck.Warrants) > 1 {
		return nil, service.NewInvalidParameterError("warrants", "must include operator when including multiple warrants")
	}

	warrantSpec := warrantCheck.Warrants[0]
	match, decisionPath, isImplicit, err := svc.Check(ctx, authInfo, CheckSpec{
		CheckWarrantSpec: warrantSpec,
		Debug:            warrantCheck.Debug,
	})
	if err != nil {
		return nil, err
	}

	if warrantCheck.Debug {
		checkResult.ProcessingTime = time.Since(start).Milliseconds()
		if len(decisionPath) > 0 {
			checkResult.DecisionPath[warrantSpec.String()] = decisionPath
		}
	}

	if match {
		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		checkResult.IsImplicit = isImplicit
		return &checkResult, nil
	}

	checkResult.Code = http.StatusForbidden
	checkResult.Result = NotAuthorized
	checkResult.IsImplicit = false
	return &checkResult, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context.
func (svc CheckService) Check(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec) (bool, []warrant.WarrantSpec, bool, error) {
	// Used to automatically append tenant context for session token w/ tenantId checks
	if authInfo != nil && authInfo.TenantId != "" {
		if warrantCheck.CheckWarrantSpec.Context == nil {
			warrantCheck.CheckWarrantSpec.Context = warrant.PolicyContext{
				"tenant": authInfo.TenantId,
			}
		} else {
			warrantCheck.CheckWarrantSpec.Context["tenant"] = authInfo.TenantId
		}
	}

	resultsC := make(chan result, 1)
	pipeline := NewPipeline(svc.checkConfig.Concurrency, svc.checkConfig.MaxConcurrency)

	checkCtx, err := svc.createCheckContext(ctx)
	if err != nil {
		return false, nil, false, err
	}
	childCtx, cancelFunc := context.WithTimeout(checkCtx, svc.checkConfig.Timeout)
	defer cancelFunc()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				resultsC <- result{
					Matched: false,
					Err:     errors.New(fmt.Sprintf("check: panic: %v", err)),
				}
			}
		}()

		svc.check(0, pipeline, childCtx, warrantCheck, make([]warrant.WarrantSpec, 0), resultsC)
	}()

	result := <-resultsC

	if result.Err != nil {
		return false, nil, false, result.Err
	}

	if result.Matched {
		return true, result.DecisionPath, len(result.DecisionPath) != 1 || result.DecisionPath[0].Relation != warrantCheck.Relation, nil
	}

	return false, nil, false, nil
}

type result struct {
	Matched      bool
	DecisionPath []warrant.WarrantSpec
	Err          error
}

func (svc CheckService) check(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled check[%d] [%s]", level, checkSpec)
		return
	default:
		start := time.Now()
		defer func() {
			log.Ctx(ctx).Debug().Msgf("exec check[%d] [%s] [%s]", level, checkSpec, time.Since(start))
		}()

		// 1. Check for direct warrant match
		matchedWarrant, err := svc.getWithPolicyMatch(ctx, checkPipeline, checkSpec.CheckWarrantSpec)
		if err != nil {
			resultC <- result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          err,
			}

			return
		}
		if matchedWarrant != nil {
			resultC <- result{
				Matched:      true,
				DecisionPath: append([]warrant.WarrantSpec{*matchedWarrant}, currentPath...),
				Err:          nil,
			}
			return
		}

		// 2. Check through indirect/group warrants
		var additionalTasks []func(execCtx context.Context, resultC chan<- result)
		additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
			svc.checkGroup(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC)
		})

		// 3. And/or defined rules for target relation
		objectTypeSpec, err := svc.objectTypeSvc.GetByTypeId(ctx, checkSpec.ObjectType)
		if err != nil {
			resultC <- result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          err,
			}
			return
		}
		if relationRule, ok := objectTypeSpec.Relations[checkSpec.Relation]; ok {
			additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
				svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, &relationRule)
			})
		}

		checkPipeline.AnyOf(ctx, resultC, additionalTasks)
	}
}

func (svc CheckService) checkGroup(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled checkGroup[%d] [%s]", level, checkSpec)
		return
	default:
		start := time.Now()
		defer func() {
			log.Ctx(ctx).Debug().Msgf("exec checkGroup[%d] [%s] [%s]", level, checkSpec, time.Since(start))
		}()

		warrants, err := svc.getMatchingSubjects(ctx, checkPipeline, checkSpec.ObjectType, checkSpec.ObjectId, checkSpec.Relation, checkSpec.Context)
		if err != nil {
			resultC <- result{
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
			resultC <- result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
			return
		}
		var additionalTasks []func(execCtx context.Context, resultC chan<- result)
		for _, w := range matchingWarrants {
			matchingWarrant := w
			additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
				svc.check(level+1, checkPipeline, execCtx, CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: matchingWarrant.Subject.ObjectType,
						ObjectId:   matchingWarrant.Subject.ObjectId,
						Relation:   matchingWarrant.Subject.Relation,
						Subject:    checkSpec.Subject,
						Context:    checkSpec.Context,
					},
					Debug: checkSpec.Debug,
				}, append([]warrant.WarrantSpec{matchingWarrant}, currentPath...), resultC)
			})
		}
		checkPipeline.AnyOf(ctx, resultC, additionalTasks)
	}
}

func (svc CheckService) checkRule(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result, rule *objecttype.RelationRule) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled checkRule[%d] [%s] [%s]", level, checkSpec, rule)
		return
	default:
		start := time.Now()
		defer func() {
			log.Ctx(ctx).Debug().Msgf("exec checkRule[%d] [%s] [%s] [%s]", level, checkSpec, rule, time.Since(start))
		}()

		warrantSpec := checkSpec.CheckWarrantSpec
		if rule == nil {
			resultC <- result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
			return
		}
		switch rule.InheritIf {
		case "":
			// No match found
			resultC <- result{
				Matched:      false,
				DecisionPath: currentPath,
				Err:          nil,
			}
		case objecttype.InheritIfAllOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, &subRule)
				})
			}
			checkPipeline.AllOf(ctx, resultC, additionalTasks)
		case objecttype.InheritIfAnyOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, &subRule)
				})
			}
			checkPipeline.AnyOf(ctx, resultC, additionalTasks)
		case objecttype.InheritIfNoneOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, &subRule)
				})
			}
			checkPipeline.NoneOf(ctx, resultC, additionalTasks)
		default:
			if rule.OfType == "" && rule.WithRelation == "" {
				svc.check(level+1, checkPipeline, ctx, CheckSpec{
					CheckWarrantSpec: CheckWarrantSpec{
						ObjectType: warrantSpec.ObjectType,
						ObjectId:   warrantSpec.ObjectId,
						Relation:   rule.InheritIf,
						Subject:    warrantSpec.Subject,
						Context:    warrantSpec.Context,
					},
					Debug: checkSpec.Debug,
				}, currentPath, resultC)
				return
			}

			matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, checkPipeline, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
			if err != nil {
				resultC <- result{
					Matched:      false,
					DecisionPath: currentPath,
					Err:          err,
				}
				return
			}
			if len(matchingWarrants) == 0 {
				resultC <- result{
					Matched:      false,
					DecisionPath: currentPath,
					Err:          nil,
				}
				return
			}
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, w := range matchingWarrants {
				matchingWarrant := w
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.check(level+1, checkPipeline, execCtx, CheckSpec{
						CheckWarrantSpec: CheckWarrantSpec{
							ObjectType: matchingWarrant.Subject.ObjectType,
							ObjectId:   matchingWarrant.Subject.ObjectId,
							Relation:   rule.InheritIf,
							Subject:    warrantSpec.Subject,
							Context:    warrantSpec.Context,
						},
						Debug: checkSpec.Debug,
					}, append([]warrant.WarrantSpec{matchingWarrant}, currentPath...), resultC)
				})
			}
			checkPipeline.AnyOf(ctx, resultC, additionalTasks)
		}
	}
}

type pipeline struct {
	serviceSemaphore chan struct{}
	subtaskSemaphore chan struct{}
}

func NewPipeline(maxServiceConcurrency int, maxSubtaskConcurrency int) *pipeline {
	return &pipeline{
		serviceSemaphore: make(chan struct{}, maxServiceConcurrency),
		subtaskSemaphore: make(chan struct{}, maxSubtaskConcurrency),
	}
}

func (p *pipeline) AcquireServiceLock() {
	p.serviceSemaphore <- struct{}{}
}

func (p *pipeline) ReleaseServiceLock() {
	<-p.serviceSemaphore
}

func (p *pipeline) AnyOf(ctx context.Context, parentResultC chan<- result, tasks []func(execCtx context.Context, resultC chan<- result)) {
	p.execTasks(ctx, parentResultC, tasks, func(res result, isLastExpected bool) (*result, bool) {
		// Short-circuit - pick this result if it's a match
		if res.Matched {
			return &res, true
		}
		// Last result AND it's not a match due to prev condition -> return not matched
		if isLastExpected {
			return &result{
				Matched:      false,
				DecisionPath: res.DecisionPath,
				Err:          nil,
			}, true
		}
		// Not a match, keep looking
		return nil, false
	})
}

func (p *pipeline) AllOf(ctx context.Context, parentResultC chan<- result, tasks []func(execCtx context.Context, resultC chan<- result)) {
	p.execTasks(ctx, parentResultC, tasks, func(res result, isLastExpected bool) (*result, bool) {
		// Short-circuit - return not matched if any sub-result is not matched
		if !res.Matched {
			return &res, true
		}
		// Last result AND it's a match due to prev condition -> return matched
		if isLastExpected {
			return &result{
				Matched:      true,
				DecisionPath: res.DecisionPath,
				Err:          nil,
			}, true
		}
		// Keep looking
		return nil, false
	})
}

func (p *pipeline) NoneOf(ctx context.Context, parentResultC chan<- result, tasks []func(execCtx context.Context, resultC chan<- result)) {
	p.execTasks(ctx, parentResultC, tasks, func(res result, isLastExpected bool) (*result, bool) {
		// Short-circuit - return not matched
		if res.Matched {
			return &result{
				Matched:      false,
				DecisionPath: res.DecisionPath,
				Err:          nil,
			}, true
		}
		// Last result AND it's not a match due to prev condition -> return matched
		if isLastExpected {
			return &result{
				Matched:      true,
				DecisionPath: res.DecisionPath,
				Err:          nil,
			}, true
		}
		// Keep looking
		return nil, false
	})
}

func (p *pipeline) execTasks(ctx context.Context, parentResultC chan<- result, tasks []func(execCtx context.Context, resultC chan<- result), checkResultFunc func(r result, isLastExpected bool) (*result, bool)) {
	childContext, childCtxCancelFunc := context.WithCancel(ctx)
	childResultC := make(chan result, len(tasks))

	go func() {
		defer func() {
			if err := recover(); err != nil {
				parentResultC <- result{
					Matched: false,
					Err:     errors.New(fmt.Sprintf("check: panic: %v", err)),
				}
			}
		}()

		// Monitor task results, short-circuit as needed
		defer childCtxCancelFunc()
		resultsReceived := 0
		for result := range childResultC {
			if result.Err != nil {
				parentResultC <- result
				return
			}
			resultsReceived++
			r, returnResult := checkResultFunc(result, resultsReceived == len(tasks))
			if returnResult {
				parentResultC <- *r
				return
			}
		}
	}()

	for _, t := range tasks {
		task := t
		// Exec each task on new goroutine unless at capacity. In that case, run task(s) locally
		select {
		case p.subtaskSemaphore <- struct{}{}:
			go func() {
				defer func() {
					if err := recover(); err != nil {
						childResultC <- result{
							Matched: false,
							Err:     errors.New(fmt.Sprintf("check: panic: %v", err)),
						}
					}
					<-p.subtaskSemaphore
				}()

				task(childContext, childResultC)
			}()
		default:
			task(childContext, childResultC)
		}
	}
}

func evalWarrantPolicy(w warrant.WarrantSpec, policyCtx warrant.PolicyContext) bool {
	policyCtxWithWarrant := make(warrant.PolicyContext)
	for k, v := range policyCtx {
		policyCtxWithWarrant[k] = v
	}
	policyCtxWithWarrant["warrant"] = w

	policyMatched, err := w.Policy.Eval(policyCtxWithWarrant)
	if err != nil {
		log.Err(err).Msgf("check: error while evaluating policy %s", w.Policy)
		return false
	}

	return policyMatched
}
