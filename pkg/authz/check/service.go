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
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CheckService struct {
	service.BaseService
	WarrantRepository warrant.WarrantRepository
	EventSvc          event.Service
	ObjectTypeSvc     *objecttype.ObjectTypeService
	WookieSvc         *wookie.WookieService
	CheckConfig       *config.CheckConfig
}

func NewService(env service.Env, warrantRepo warrant.WarrantRepository, eventSvc event.Service, objectTypeSvc *objecttype.ObjectTypeService, wookieSvc *wookie.WookieService, checkConfig *config.CheckConfig) *CheckService {
	return &CheckService{
		BaseService:       service.NewBaseService(env),
		WarrantRepository: warrantRepo,
		EventSvc:          eventSvc,
		ObjectTypeSvc:     objectTypeSvc,
		WookieSvc:         wookieSvc,
		CheckConfig:       checkConfig,
	}
}

func (svc CheckService) getWithPolicyMatch(ctx context.Context, checkPipeline *pipeline, spec CheckWarrantSpec) (*warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	warrants, err := svc.WarrantRepository.GetAllMatchingObjectRelationAndSubject(ctx, spec.ObjectType, spec.ObjectId, spec.Relation, spec.Subject.ObjectType, spec.Subject.ObjectId, spec.Subject.Relation)
	if err != nil || len(warrants) == 0 {
		return nil, err
	}

	// if a warrant without a policy is found, match it
	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			return warrant.ToWarrantSpec(), nil
		}
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() != "" {
			if policyMatched := evalWarrantPolicy(warrant, spec.Context); policyMatched {
				return warrant.ToWarrantSpec(), nil
			}
		}
	}

	return nil, nil
}

func (svc CheckService) getMatchingSubjects(ctx context.Context, checkPipeline *pipeline, objectTypeMap *objecttype.ObjectTypeMap, objectType string, objectId string, relation string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := objectTypeMap.GetByTypeId(objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	warrants, err := svc.WarrantRepository.GetAllMatchingObjectAndRelation(
		ctx,
		objectType,
		objectId,
		relation,
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
		} else {
			if policyMatched := evalWarrantPolicy(warrant, checkCtx); policyMatched {
				warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
			}
		}
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) getMatchingSubjectsBySubjectType(ctx context.Context, checkPipeline *pipeline, objectTypeMap *objecttype.ObjectTypeMap, objectType string,
	objectId string, relation string, subjectType string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	checkPipeline.AcquireServiceLock()
	defer checkPipeline.ReleaseServiceLock()

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := objectTypeMap.GetByTypeId(objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	warrants, err := svc.WarrantRepository.GetAllMatchingObjectAndRelationBySubjectType(
		ctx,
		objectType,
		objectId,
		relation,
		subjectType,
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		if warrant.GetPolicy() == "" {
			warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
		} else {
			if policyMatched := evalWarrantPolicy(warrant, checkCtx); policyMatched {
				warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
			}
		}
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) CheckMany(ctx context.Context, authInfo *service.AuthInfo, warrantCheck *CheckManySpec) (*CheckResultSpec, *wookie.Token, error) {
	start := time.Now().UTC()
	if warrantCheck.Op != "" && warrantCheck.Op != objecttype.InheritIfAllOf && warrantCheck.Op != objecttype.InheritIfAnyOf {
		return nil, nil, service.NewInvalidParameterError("op", "must be either anyOf or allOf")
	}

	var checkResult CheckResultSpec
	checkResult.DecisionPath = make(map[string][]warrant.WarrantSpec, 0)

	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		if warrantCheck.Op == objecttype.InheritIfAllOf {
			var processingTime int64
			for _, warrantSpec := range warrantCheck.Warrants {
				match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
					CheckWarrantSpec: warrantSpec,
					Debug:            warrantCheck.Debug,
				})
				if err != nil {
					return err
				}

				if warrantCheck.Debug {
					checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
					if len(decisionPath) > 0 {
						checkResult.DecisionPath[warrantSpec.String()] = decisionPath
					}
				}

				var eventMeta map[string]interface{}
				if warrantSpec.Context != nil {
					eventMeta = make(map[string]interface{})
					eventMeta["context"] = warrantSpec.Context
				}

				if !match {
					err = svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusForbidden
					checkResult.Result = NotAuthorized
					return nil
				}

				err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
				if err != nil {
					return err
				}
			}

			checkResult.Code = http.StatusOK
			checkResult.Result = Authorized
			return nil
		}

		if warrantCheck.Op == objecttype.InheritIfAnyOf {
			var processingTime int64
			for _, warrantSpec := range warrantCheck.Warrants {
				match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
					CheckWarrantSpec: warrantSpec,
					Debug:            warrantCheck.Debug,
				})
				if err != nil {
					return err
				}

				if warrantCheck.Debug {
					checkResult.ProcessingTime = processingTime + time.Since(start).Milliseconds()
					if len(decisionPath) > 0 {
						checkResult.DecisionPath[warrantSpec.String()] = decisionPath
					}
				}

				var eventMeta map[string]interface{}
				if warrantSpec.Context != nil {
					eventMeta = make(map[string]interface{})
					eventMeta["context"] = warrantSpec.Context
				}

				if match {
					err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusOK
					checkResult.Result = Authorized
					return nil
				}

				if !match {
					err := svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}
				}
			}

			checkResult.Code = http.StatusForbidden
			checkResult.Result = NotAuthorized
			return nil
		}

		if len(warrantCheck.Warrants) > 1 {
			return service.NewInvalidParameterError("warrants", "must include operator when including multiple warrants")
		}

		warrantSpec := warrantCheck.Warrants[0]
		match, decisionPath, _, err := svc.Check(wkCtx, authInfo, CheckSpec{
			CheckWarrantSpec: warrantSpec,
			Debug:            warrantCheck.Debug,
		})
		if err != nil {
			return err
		}

		if warrantCheck.Debug {
			checkResult.ProcessingTime = time.Since(start).Milliseconds()
			if len(decisionPath) > 0 {
				checkResult.DecisionPath[warrantSpec.String()] = decisionPath
			}
		}

		var eventMeta map[string]interface{}
		if warrantSpec.Context != nil {
			eventMeta = make(map[string]interface{})
			eventMeta["context"] = warrantSpec.Context
		}

		if match {
			err = svc.EventSvc.TrackAccessAllowedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
			if err != nil {
				return err
			}

			checkResult.Code = http.StatusOK
			checkResult.Result = Authorized
			return nil
		}

		err = svc.EventSvc.TrackAccessDeniedEvent(wkCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
		if err != nil {
			return err
		}

		checkResult.Code = http.StatusForbidden
		checkResult.Result = NotAuthorized
		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return &checkResult, newWookie, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context
func (svc CheckService) Check(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec) (bool, []warrant.WarrantSpec, *wookie.Token, error) {
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

	// Fetch object types upfront
	typesMap, _, err := svc.ObjectTypeSvc.GetTypeMap(ctx)
	if err != nil {
		return false, nil, nil, err
	}

	// TODO: Should do wookieSafeRead
	resultsC := make(chan result, 1)
	pipeline := NewPipeline(svc.CheckConfig.Concurrency, svc.CheckConfig.MaxConcurrency)

	childCtx, cancelFunc := context.WithTimeout(ctx, svc.CheckConfig.Timeout)
	defer cancelFunc()

	go func() {
		svc.check(0, pipeline, childCtx, warrantCheck, make([]warrant.WarrantSpec, 0), resultsC, typesMap)
	}()

	result := <-resultsC

	if result.Err != nil {
		return false, nil, nil, result.Err
	}

	if result.Matched {
		return true, result.DecisionPath, nil, nil
	}

	return false, nil, nil, nil
}

type result struct {
	Matched      bool
	DecisionPath []warrant.WarrantSpec
	Err          error
}

func (svc CheckService) check(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result, typesMap *objecttype.ObjectTypeMap) {
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
			svc.checkGroup(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, typesMap)
		})

		// 3. And/or defined rules for target relation
		objectTypeSpec, err := typesMap.GetByTypeId(checkSpec.ObjectType)
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
				svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, typesMap, &relationRule)
			})
		}

		checkPipeline.AnyOf(ctx, resultC, additionalTasks)
	}
}

func (svc CheckService) checkGroup(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result, typesMap *objecttype.ObjectTypeMap) {
	select {
	case <-ctx.Done():
		log.Ctx(ctx).Debug().Msgf("canceled checkGroup[%d] [%s]", level, checkSpec)
		return
	default:
		start := time.Now()
		defer func() {
			log.Ctx(ctx).Debug().Msgf("exec checkGroup[%d] [%s] [%s]", level, checkSpec, time.Since(start))
		}()

		warrants, err := svc.getMatchingSubjects(ctx, checkPipeline, typesMap, checkSpec.ObjectType, checkSpec.ObjectId, checkSpec.Relation, checkSpec.Context)
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
				}, append([]warrant.WarrantSpec{matchingWarrant}, currentPath...), resultC, typesMap)
			})
		}
		checkPipeline.AnyOf(ctx, resultC, additionalTasks)
	}
}

func (svc CheckService) checkRule(level int, checkPipeline *pipeline, ctx context.Context, checkSpec CheckSpec, currentPath []warrant.WarrantSpec, resultC chan<- result, typesMap *objecttype.ObjectTypeMap, rule *objecttype.RelationRule) {
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
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, typesMap, &subRule)
				})
			}
			checkPipeline.AllOf(ctx, resultC, additionalTasks)
		case objecttype.InheritIfAnyOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, typesMap, &subRule)
				})
			}
			checkPipeline.AnyOf(ctx, resultC, additionalTasks)
		case objecttype.InheritIfNoneOf:
			var additionalTasks []func(execCtx context.Context, resultC chan<- result)
			for _, r := range rule.Rules {
				subRule := r
				additionalTasks = append(additionalTasks, func(execCtx context.Context, resultC chan<- result) {
					svc.checkRule(level+1, checkPipeline, execCtx, checkSpec, currentPath, resultC, typesMap, &subRule)
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
				}, currentPath, resultC, typesMap)
				return
			}

			matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, checkPipeline, typesMap, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
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
					}, append([]warrant.WarrantSpec{matchingWarrant}, currentPath...), resultC, typesMap)
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
					<-p.subtaskSemaphore
				}()
				task(childContext, childResultC)
			}()
		default:
			task(childContext, childResultC)
		}
	}
}

func evalWarrantPolicy(w warrant.Model, policyCtx warrant.PolicyContext) bool {
	policyCtxWithWarrant := make(warrant.PolicyContext)
	for k, v := range policyCtx {
		policyCtxWithWarrant[k] = v
	}
	policyCtxWithWarrant["warrant"] = w

	policyMatched, err := w.GetPolicy().Eval(policyCtxWithWarrant)
	if err != nil {
		log.Err(err).Msgf("Error while evaluating policy %s", w.GetPolicy())
		return false
	}

	return policyMatched
}
