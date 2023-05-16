package authz

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wntContext "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CheckService struct {
	service.BaseService
	WarrantRepository warrant.WarrantRepository
	EventSvc          event.EventService
	CtxSvc            wntContext.ContextService
	ObjectTypeSvc     objecttype.ObjectTypeService
}

func NewService(env service.Env, warrantRepo warrant.WarrantRepository, ctxSvc wntContext.ContextService, eventSvc event.EventService, objectTypeSvc objecttype.ObjectTypeService) CheckService {
	return CheckService{
		BaseService:       service.NewBaseService(env),
		WarrantRepository: warrantRepo,
		CtxSvc:            ctxSvc,
		EventSvc:          eventSvc,
		ObjectTypeSvc:     objectTypeSvc,
	}
}

func (svc CheckService) getWithContextMatch(ctx context.Context, spec warrant.WarrantSpec) (*warrant.WarrantSpec, error) {
	warrant, err := svc.WarrantRepository.GetWithContextMatch(ctx, spec.ObjectType, spec.ObjectId, spec.Relation, spec.Subject.ObjectType, spec.Subject.ObjectId, spec.Subject.Relation, spec.Context.ToHash())
	if err != nil || warrant == nil {
		return nil, err
	}

	contextSetSpec, err := svc.CtxSvc.ListByWarrantId(ctx, []int64{warrant.GetID()})
	if err != nil {
		return nil, err
	}

	warrantSpec := warrant.ToWarrantSpec()
	warrantSpec.Context = contextSetSpec[warrant.GetID()]
	return warrantSpec, nil
}

func (svc CheckService) getMatchingSubjects(ctx context.Context, objectType string, objectId string, relation string, wntCtx wntContext.ContextSetSpec) ([]warrant.WarrantSpec, error) {
	log.Debug().Msgf("Getting matching subjects for %s:%s#%s@___%s", objectType, objectId, relation, wntCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
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
		wntCtx.ToHash(),
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) getMatchingSubjectsBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string, wntCtx wntContext.ContextSetSpec) ([]warrant.WarrantSpec, error) {
	log.Debug().Msgf("Getting matching subjects for %s:%s#%s@%s:___%s", objectType, objectId, relation, subjectType, wntCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	objectTypeSpec, err := svc.ObjectTypeSvc.GetByTypeId(ctx, objectType)
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
		wntCtx.ToHash(),
	)
	if err != nil {
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	if err != nil {
		return warrantSpecs, err
	}

	return warrantSpecs, nil
}

func (svc CheckService) checkRule(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec, rule *objecttype.RelationRule) (match bool, decisionPath []warrant.WarrantSpec, err error) {
	warrantSpec := warrantCheck.WarrantSpec
	if rule == nil {
		return false, decisionPath, nil
	}

	switch rule.InheritIf {
	case "":
		// No match found
		return false, decisionPath, nil
	case objecttype.InheritIfAllOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if !isMatch {
				return false, decisionPath, nil
			}
		}

		return true, decisionPath, nil
	case objecttype.InheritIfAnyOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return true, decisionPath, nil
			}
		}

		return false, decisionPath, nil
	case objecttype.InheritIfNoneOf:
		for _, r := range rule.Rules {
			isMatch, matchedPath, err := svc.checkRule(ctx, authInfo, warrantCheck, &r)
			if err != nil {
				return false, decisionPath, err
			}

			decisionPath = append(decisionPath, matchedPath...)
			if isMatch {
				return false, decisionPath, nil
			}
		}

		return true, decisionPath, nil
	default:
		if rule.OfType == "" && rule.WithRelation == "" {
			return svc.Check(ctx, authInfo, CheckSpec{
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
				WarrantSpec: warrant.WarrantSpec{
					ObjectType: warrantSpec.ObjectType,
					ObjectId:   warrantSpec.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
			})
		}

		matchingWarrants, err := svc.getMatchingSubjectsBySubjectType(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
		if err != nil {
			return false, decisionPath, err
		}

		for _, matchingWarrant := range matchingWarrants {
			match, decisionPath, err := svc.Check(ctx, authInfo, CheckSpec{
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
				WarrantSpec: warrant.WarrantSpec{
					ObjectType: matchingWarrant.Subject.ObjectType,
					ObjectId:   matchingWarrant.Subject.ObjectId,
					Relation:   rule.InheritIf,
					Subject:    warrantSpec.Subject,
					Context:    warrantSpec.Context,
				},
			})
			if err != nil {
				return false, decisionPath, err
			}

			if match {
				return true, decisionPath, nil
			}
		}

		return false, decisionPath, nil
	}
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
		for _, warrantSpec := range warrantCheck.Warrants {
			match, decisionPath, err := svc.Check(ctx, authInfo, CheckSpec{
				WarrantSpec:    warrantSpec,
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
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

			if !match {
				err = svc.EventSvc.TrackAccessDeniedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
				if err != nil {
					return nil, err
				}

				checkResult.Code = http.StatusForbidden
				checkResult.Result = NotAuthorized
				return &checkResult, nil
			}

			err = svc.EventSvc.TrackAccessAllowedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
			if err != nil {
				return nil, err
			}
		}

		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		return &checkResult, nil
	}

	if warrantCheck.Op == objecttype.InheritIfAnyOf {
		var processingTime int64
		for _, warrantSpec := range warrantCheck.Warrants {
			match, decisionPath, err := svc.Check(ctx, authInfo, CheckSpec{
				WarrantSpec:    warrantSpec,
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
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
				err = svc.EventSvc.TrackAccessAllowedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
				if err != nil {
					return nil, err
				}

				checkResult.Code = http.StatusOK
				checkResult.Result = Authorized
				return &checkResult, nil
			}

			if !match {
				err := svc.EventSvc.TrackAccessDeniedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
				if err != nil {
					return nil, err
				}
			}
		}

		checkResult.Code = http.StatusForbidden
		checkResult.Result = NotAuthorized
		return &checkResult, nil
	}

	if len(warrantCheck.Warrants) > 1 {
		return nil, service.NewInvalidParameterError("warrants", "must include operator when including multiple warrants")
	}

	warrantSpec := warrantCheck.Warrants[0]
	match, decisionPath, err := svc.Check(ctx, authInfo, CheckSpec{
		WarrantSpec:    warrantSpec,
		ConsistentRead: warrantCheck.ConsistentRead,
		Debug:          warrantCheck.Debug,
	})
	if err != nil {
		return nil, err
	}

	if match {
		err = svc.EventSvc.TrackAccessAllowedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
		if err != nil {
			return nil, err
		}

		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		return &checkResult, nil
	}

	err = svc.EventSvc.TrackAccessDeniedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
	if err != nil {
		return nil, err
	}

	checkResult.Code = http.StatusForbidden
	checkResult.Result = NotAuthorized
	if warrantCheck.Debug {
		checkResult.ProcessingTime = time.Since(start).Milliseconds()
		if len(decisionPath) > 0 {
			checkResult.DecisionPath[warrantSpec.String()] = decisionPath
		}
	}

	return &checkResult, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context
func (svc CheckService) Check(ctx context.Context, authInfo *service.AuthInfo, warrantCheck CheckSpec) (match bool, decisionPath []warrant.WarrantSpec, err error) {
	log.Debug().Msgf("Checking for warrant %s", warrantCheck.String())

	// Used to automatically append tenant context for session token w/ tenantId checks
	if authInfo != nil && authInfo.TenantId != "" {
		svc.appendTenantContext(&warrantCheck, authInfo.TenantId)
	}

	// Check for direct warrant match -> doc:readme#viewer@[10]
	matchedWarrant, err := svc.getWithContextMatch(ctx, warrantCheck.WarrantSpec)
	if err != nil {
		return false, decisionPath, err
	}

	if matchedWarrant != nil {
		return true, []warrant.WarrantSpec{{
			ObjectType: matchedWarrant.ObjectType,
			ObjectId:   matchedWarrant.ObjectId,
			Relation:   matchedWarrant.Relation,
			Subject:    matchedWarrant.Subject,
			Context:    matchedWarrant.Context,
			CreatedAt:  matchedWarrant.CreatedAt,
		}}, nil
	}

	// Check against indirectly related warrants
	matchingWarrants, err := svc.getMatchingSubjects(ctx, warrantCheck.ObjectType, warrantCheck.ObjectId, warrantCheck.Relation, warrantCheck.Context)
	if err != nil {
		return false, decisionPath, err
	}

	for _, matchingWarrant := range matchingWarrants {
		if matchingWarrant.Subject.Relation == "" {
			continue
		}

		match, decisionPath, err := svc.Check(ctx, authInfo, CheckSpec{
			ConsistentRead: warrantCheck.ConsistentRead,
			Debug:          warrantCheck.Debug,
			WarrantSpec: warrant.WarrantSpec{
				ObjectType: matchingWarrant.Subject.ObjectType,
				ObjectId:   matchingWarrant.Subject.ObjectId,
				Relation:   matchingWarrant.Subject.Relation,
				Subject:    warrantCheck.Subject,
				Context:    warrantCheck.Context,
			},
		})
		if err != nil {
			return false, decisionPath, err
		}

		if match {
			return true, decisionPath, nil
		}
	}

	// Attempt to match against defined rules for target relation
	objectTypeSpec, err := svc.ObjectTypeSvc.GetByTypeId(ctx, warrantCheck.ObjectType)
	if err != nil {
		return false, decisionPath, err
	}

	relationRule := objectTypeSpec.Relations[warrantCheck.Relation]
	match, decisionPath, err = svc.checkRule(ctx, authInfo, warrantCheck, &relationRule)
	if err != nil {
		return false, decisionPath, err
	}

	if match {
		return true, decisionPath, nil
	}

	return false, decisionPath, nil
}

func (svc CheckService) appendTenantContext(warrantCheck *CheckSpec, tenantId string) {
	if warrantCheck.WarrantSpec.Context == nil {
		warrantCheck.WarrantSpec.Context = wntContext.ContextSetSpec{
			"tenant": tenantId,
		}
	} else {
		warrantCheck.WarrantSpec.Context["tenant"] = tenantId
	}
}
