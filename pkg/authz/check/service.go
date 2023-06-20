package authz

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CheckService struct {
	service.BaseService
	WarrantRepository warrant.WarrantRepository
	EventSvc          event.EventService
	ObjectTypeSvc     objecttype.ObjectTypeService
	WookieService     wookie.WookieService
}

func NewService(env service.Env, warrantRepo warrant.WarrantRepository, eventSvc event.EventService, objectTypeSvc objecttype.ObjectTypeService, wookieService wookie.WookieService) CheckService {
	return CheckService{
		BaseService:       service.NewBaseService(env),
		WarrantRepository: warrantRepo,
		EventSvc:          eventSvc,
		ObjectTypeSvc:     objectTypeSvc,
		WookieService:     wookieService,
	}
}

func (svc CheckService) getWithPolicyMatch(ctx context.Context, spec CheckWarrantSpec) (*warrant.WarrantSpec, error) {
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

// TODO: change/fix latestWookie handling
func (svc CheckService) getMatchingSubjects(ctx context.Context, objectType string, objectId string, relation string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	log.Ctx(ctx).Debug().Msgf("Getting matching subjects for %s:%s#%s@___%s", objectType, objectId, relation, checkCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
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

func (svc CheckService) getMatchingSubjectsBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string, checkCtx warrant.PolicyContext) ([]warrant.WarrantSpec, error) {
	log.Ctx(ctx).Debug().Msgf("Getting matching subjects for %s:%s#%s@%s:___%s", objectType, objectId, relation, subjectType, checkCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
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

func (svc CheckService) checkRule(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec, rule *objecttype.RelationRule) (match bool, decisionPath []warrant.WarrantSpec, latestWookie *wookie.Token, err error) {
	warrantSpec := warrantCheck.CheckWarrantSpec
	if rule == nil {
		return false, decisionPath, latestWookie, nil
	}

	switch rule.InheritIf {
	case "":
		// No match found
		return false, decisionPath, latestWookie, nil
	case objecttype.InheritIfAllOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, tok, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, tok, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if !isMatch {
				return false, decisionPath, latestWookie, nil
			}
		}

		return true, decisionPath, latestWookie, nil
	case objecttype.InheritIfAnyOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, tok, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, tok, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return true, decisionPath, latestWookie, nil
			}
		}

		return false, decisionPath, latestWookie, nil
	case objecttype.InheritIfNoneOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, tok, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, tok, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return false, decisionPath, latestWookie, nil
			}
		}

		return true, decisionPath, latestWookie, nil
	default:
		if rule.OfType == "" && rule.WithRelation == "" {
			return svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: warrantSpec.ObjectType,
					ObjectId:   warrantSpec.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
				Debug: warrantCheck.Debug,
			})
		}

		matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
		if err != nil {
			return false, decisionPath, latestWookie, err
		}

		for _, matchingWarrant := range matchingWarrants {
			match, decisionPath, tok, err := svc.Check(ctx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
				Debug: warrantCheck.Debug,
			})
			if err != nil {
				return false, decisionPath, tok, err
			}

			if match {
				decisionPath = append(decisionPath, matchingWarrant)
				return true, decisionPath, tok, nil
			}
		}

		return false, decisionPath, latestWookie, nil
	}
}

func (svc CheckService) CheckMany(ctx context.Context, authInfo *service.AuthInfo, warrantCheck *CheckManySpec) (*CheckResultSpec, *wookie.Token, error) {
	start := time.Now().UTC()
	if warrantCheck.Op != "" && warrantCheck.Op != objecttype.InheritIfAllOf && warrantCheck.Op != objecttype.InheritIfAnyOf {
		return nil, nil, service.NewInvalidParameterError("op", "must be either anyOf or allOf")
	}

	var latestWookie *wookie.Token
	var checkResult CheckResultSpec
	checkResult.DecisionPath = make(map[string][]warrant.WarrantSpec, 0)

	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		wookieCtx, token, err := svc.WookieService.GetWookieContext(connCtx)
		if err != nil {
			return err
		}
		latestWookie = token

		if warrantCheck.Op == objecttype.InheritIfAllOf {
			var processingTime int64
			for _, warrantSpec := range warrantCheck.Warrants {
				match, decisionPath, _, err := svc.Check(wookieCtx, authInfo, CheckSpec{
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
					err = svc.EventSvc.TrackAccessDeniedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusForbidden
					checkResult.Result = NotAuthorized
					return nil
				}

				err = svc.EventSvc.TrackAccessAllowedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
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
				match, decisionPath, _, err := svc.Check(wookieCtx, authInfo, CheckSpec{
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
					err = svc.EventSvc.TrackAccessAllowedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
					if err != nil {
						return err
					}

					checkResult.Code = http.StatusOK
					checkResult.Result = Authorized
					return nil
				}

				if !match {
					err := svc.EventSvc.TrackAccessDeniedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
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
		match, decisionPath, _, err := svc.Check(wookieCtx, authInfo, CheckSpec{
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
			err = svc.EventSvc.TrackAccessAllowedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
			if err != nil {
				return err
			}

			checkResult.Code = http.StatusOK
			checkResult.Result = Authorized
			return nil
		}

		err = svc.EventSvc.TrackAccessDeniedEvent(wookieCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
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
	return &checkResult, latestWookie, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context
func (svc CheckService) Check(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec) (match bool, decisionPath []warrant.WarrantSpec, latestWookie *wookie.Token, err error) {
	log.Ctx(ctx).Debug().Msgf("Checking for warrant %s", warrantCheck.String())

	// Used to automatically append tenant context for session token w/ tenantId checks
	if authInfo != nil && authInfo.TenantId != "" {
		svc.appendTenantContext(&warrantCheck, authInfo.TenantId)
	}

	matched := false
	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		wookieCtx, token, err := svc.WookieService.GetWookieContext(connCtx)
		if err != nil {
			matched = false
			return err
		}
		latestWookie = token

		// Check for direct warrant match -> doc:readme#viewer@[10]
		matchedWarrant, err := svc.getWithPolicyMatch(wookieCtx, warrantCheck.CheckWarrantSpec)
		if err != nil {
			matched = false
			return err
		}

		if matchedWarrant != nil {
			// TODO: shouldn't this append to decisionPath?
			// return true, []warrant.WarrantSpec{*matchedWarrant}, nil
			matched = true
			decisionPath = []warrant.WarrantSpec{*matchedWarrant}
			return nil
		}

		// Check against indirectly related warrants
		matchingWarrants, err := svc.getMatchingSubjects(wookieCtx, warrantCheck.ObjectType, warrantCheck.ObjectId, warrantCheck.Relation, warrantCheck.Context)
		if err != nil {
			matched = false
			return err
		}

		for _, matchingWarrant := range matchingWarrants {
			if matchingWarrant.Subject.Relation == "" {
				continue
			}

			// TODO: is this correct?
			// match, decisionPath, err = svc.Check(wookieCtx, authInfo, CheckSpec{
			match, decisionPath, latestWookie, err = svc.Check(wookieCtx, authInfo, CheckSpec{
				CheckWarrantSpec: CheckWarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   matchingWarrant.Subject.Relation,
					Subject:    warrantCheck.Subject,
					Context:    warrantCheck.Context,
				},
				Debug: warrantCheck.Debug,
			})
			if err != nil {
				matched = false
				return err
			}

			if match {
				// TODO: is this correct?
				decisionPath = append(decisionPath, matchingWarrant)
				//decisionPath = path
				matched = true
				return nil
			}
		}

		// Attempt to match against defined rules for target relation
		objectTypeSpec, _, err := svc.ObjectTypeSvc.GetByTypeId(wookieCtx, warrantCheck.ObjectType)
		if err != nil {
			matched = false
			return err
		}

		relationRule := objectTypeSpec.Relations[warrantCheck.Relation]
		match, decisionPath, _, err = svc.checkRule(wookieCtx, authInfo, warrantCheck, &relationRule)
		if err != nil {
			matched = false
			return err
		}

		if match {
			matched = true
			return nil
		}

		matched = false
		return nil
	})
	if e != nil {
		return false, decisionPath, latestWookie, e
	}
	return matched, decisionPath, latestWookie, nil
}

func (svc CheckService) appendTenantContext(warrantCheck *CheckSpec, tenantId string) {
	if warrantCheck.CheckWarrantSpec.Context == nil {
		warrantCheck.CheckWarrantSpec.Context = warrant.PolicyContext{
			"tenant": tenantId,
		}
	} else {
		warrantCheck.CheckWarrantSpec.Context["tenant"] = tenantId
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
