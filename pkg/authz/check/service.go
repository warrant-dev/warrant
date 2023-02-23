package authz

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	wntContext "github.com/warrant-dev/warrant/pkg/authz/context"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CheckService struct {
	service.BaseService
	objectTypeMap map[string]*objecttype.ObjectTypeSpec
}

func NewService(env service.Env) CheckService {
	return CheckService{
		BaseService:   service.NewBaseService(env),
		objectTypeMap: make(map[string]*objecttype.ObjectTypeSpec),
	}
}

func (svc CheckService) getWithContextMatch(ctx context.Context, spec warrant.WarrantSpec) (*warrant.WarrantSpec, error) {
	warrantRepository, err := warrant.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrant, err := warrantRepository.GetWithContextMatch(ctx, spec.ObjectType, spec.ObjectId, spec.Relation, spec.Subject.ObjectType, spec.Subject.ObjectId, spec.Subject.Relation, spec.Context.ToHash())
	if err != nil || warrant == nil {
		return nil, err
	}

	contextRepository, err := wntContext.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrant.Context, err = contextRepository.ListByWarrantId(ctx, []int64{warrant.ID})
	if err != nil {
		return nil, err
	}

	return warrant.ToWarrantSpec(), nil
}

func (svc CheckService) getMatchingSubjects(ctx context.Context, objectType string, objectId string, relation string, subjectType string, wntCtx wntContext.ContextSetSpec) ([]warrant.WarrantSpec, error) {
	log.Debug().Msgf("Getting matching subjects for %s:%s#%s@%s:___%s", objectType, objectId, relation, subjectType, wntCtx)

	warrantSpecs := make([]warrant.WarrantSpec, 0)
	warrantRepository, err := warrant.NewRepository(svc.Env().DB())
	if err != nil {
		return warrantSpecs, err
	}

	objectTypeSpec, err := svc.getObjectType(ctx, objectType)
	if err != nil {
		return warrantSpecs, err
	}

	if _, ok := objectTypeSpec.Relations[relation]; !ok {
		return warrantSpecs, nil
	}

	warrants, err := warrantRepository.GetAllMatchingObjectAndRelation(
		ctx,
		objectType,
		objectId,
		relation,
		subjectType,
		wntCtx.ToHash(),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Error fetching warrants for object")
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	if err != nil {
		return warrantSpecs, err
	}

	warrants, err = warrantRepository.GetAllMatchingWildcard(
		ctx,
		objectType,
		objectId,
		relation,
		wntCtx.ToHash(),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Error fetching warrants matching wildcard")
		return warrantSpecs, err
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, *warrant.ToWarrantSpec())
	}

	return warrantSpecs, nil
}

func (svc CheckService) checkRule(ctx context.Context, warrantCheck CheckSpec, rule *objecttype.RelationRule) (match bool, decisionPath []warrant.WarrantSpec, err error) {
	decisionPath = make([]warrant.WarrantSpec, 0)
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
			isMatch, matchedPath, err := svc.checkRule(ctx, warrantCheck, &r)
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
			isMatch, matchedPath, err := svc.checkRule(ctx, warrantCheck, &r)
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
			isMatch, matchedPath, err := svc.checkRule(ctx, warrantCheck, &r)
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
			return svc.Check(ctx, CheckSpec{
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

		matchingWarrants, err := svc.getMatchingSubjects(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, rule.WithRelation, rule.OfType, warrantSpec.Context)
		if err != nil {
			return false, decisionPath, err
		}

		for _, matchingWarrant := range matchingWarrants {
			match, decisionPath, err := svc.Check(ctx, CheckSpec{
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

func (svc CheckService) CheckMany(ctx context.Context, warrantCheck *CheckManySpec) (*CheckResultSpec, error) {
	start := time.Now()
	if warrantCheck.Op != "" && warrantCheck.Op != objecttype.InheritIfAllOf && warrantCheck.Op != objecttype.InheritIfAnyOf {
		return nil, service.NewInvalidParameterError("op", "must be either anyOf or allOf")
	}

	var checkResult CheckResultSpec
	if warrantCheck.Op == objecttype.InheritIfAllOf {
		for _, warrantSpec := range warrantCheck.Warrants {
			match, _, err := svc.Check(ctx, CheckSpec{
				WarrantSpec:    warrantSpec,
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
			})
			if err != nil {
				return nil, err
			}

			if !match {
				// eventsService.TrackAsync(events.AuthorizeFail, warrantCheck.ToMap())
				checkResult.Code = http.StatusForbidden
				checkResult.Result = NotAuthorized
				if warrantCheck.Debug {
					checkResult.ProcessingTime = time.Since(start).Milliseconds()
					// checkResult.DecisionPath = warrantPath
				}

				return &checkResult, nil
			}
		}

		// eventsService.TrackAsync(events.AuthorizeSuccess, warrantCheck.ToMap())
		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		if warrantCheck.Debug {
			checkResult.ProcessingTime = time.Since(start).Milliseconds()
			// checkResult.DecisionPath = warrantPath
		}

		return &checkResult, nil
	}

	if warrantCheck.Op == objecttype.InheritIfAnyOf {
		for _, warrantSpec := range warrantCheck.Warrants {
			match, _, err := svc.Check(ctx, CheckSpec{
				WarrantSpec:    warrantSpec,
				ConsistentRead: warrantCheck.ConsistentRead,
				Debug:          warrantCheck.Debug,
			})
			if err != nil {
				return nil, err
			}

			if match {
				// eventsService.TrackAsync(events.AuthorizeSuccess, warrantCheck.ToMap())
				checkResult.Code = http.StatusOK
				checkResult.Result = Authorized
				if warrantCheck.Debug {
					checkResult.ProcessingTime = time.Since(start).Milliseconds()
					// checkResult.DecisionPath = warrantPath
				}

				return &checkResult, nil
			}
		}

		// eventsService.TrackAsync(events.AuthorizeFail, warrantCheck.ToMap())
		checkResult.Code = http.StatusForbidden
		checkResult.Result = NotAuthorized
		if warrantCheck.Debug {
			checkResult.ProcessingTime = time.Since(start).Milliseconds()
			// checkResult.DecisionPath = warrantPath
		}

		return &checkResult, nil
	}

	if len(warrantCheck.Warrants) > 1 {
		return nil, service.NewInvalidParameterError("warrants", "must include operator when including multiple warrants")
	}

	warrantSpec := warrantCheck.Warrants[0]
	match, warrantPath, err := svc.Check(ctx, CheckSpec{
		WarrantSpec:    warrantSpec,
		ConsistentRead: warrantCheck.ConsistentRead,
		Debug:          warrantCheck.Debug,
	})
	if err != nil {
		return nil, err
	}

	if match {
		// eventsService.TrackAsync(events.AuthorizeSuccess, warrantCheck.ToMap())
		checkResult.Code = http.StatusOK
		checkResult.Result = Authorized
		if warrantCheck.Debug {
			checkResult.ProcessingTime = time.Since(start).Milliseconds()
			checkResult.DecisionPath = warrantPath
		}

		return &checkResult, nil
	}

	// eventsService.TrackAsync(events.AuthorizeFail, warrantSpec.ToMap())
	checkResult.Code = http.StatusForbidden
	checkResult.Result = NotAuthorized
	if warrantCheck.Debug {
		checkResult.ProcessingTime = time.Since(start).Milliseconds()
		checkResult.DecisionPath = warrantPath
	}

	return &checkResult, nil
}

// Check returns true if the subject has a warrant (explicitly or implicitly) for given objectType:objectId#relation and context
func (svc CheckService) Check(ctx context.Context, warrantCheck CheckSpec) (match bool, decisionPath []warrant.WarrantSpec, err error) {
	log.Debug().Msgf("Checking for warrant %s", warrantCheck.String())

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
		}}, nil
	}

	// Check against indirectly related warrants
	matchingWarrants, err := svc.getMatchingSubjects(ctx, warrantCheck.ObjectType, warrantCheck.ObjectId, warrantCheck.Relation, "%", warrantCheck.Context)
	if err != nil {
		return false, decisionPath, err
	}

	for _, matchingWarrant := range matchingWarrants {
		if matchingWarrant.Subject.Relation == "" {
			continue
		}

		match, decisionPath, err := svc.Check(ctx, CheckSpec{
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
	objectTypeSpec, err := svc.getObjectType(ctx, warrantCheck.ObjectType)
	if err != nil {
		return false, decisionPath, err
	}

	relationRule := objectTypeSpec.Relations[warrantCheck.Relation]
	match, decisionPath, err = svc.checkRule(ctx, warrantCheck, &relationRule)
	if err != nil {
		return false, decisionPath, err
	}

	if match {
		return true, decisionPath, nil
	}

	return false, decisionPath, nil
}

func (svc CheckService) getObjectType(ctx context.Context, objectType string) (*objecttype.ObjectTypeSpec, error) {
	if objectTypeSpec, ok := svc.objectTypeMap[objectType]; ok {
		return objectTypeSpec, nil
	}

	objectTypeSpec, err := objecttype.NewService(svc.Env()).GetByTypeId(ctx, objectType)
	if err != nil {
		return nil, err
	}

	svc.objectTypeMap[objectType] = objectTypeSpec
	return objectTypeSpec, nil
}
