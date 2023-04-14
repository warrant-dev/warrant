package authz

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeUser = "user"

type UserService struct {
	service.BaseService
	repo      UserRepository
	eventSvc  event.EventService
	objectSvc object.ObjectService
}

func NewService(env service.Env, repo UserRepository, eventSvc event.EventService, objectSvc object.ObjectService) UserService {
	return UserService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
		objectSvc:   objectSvc,
	}
}

func (svc UserService) Create(ctx context.Context, userSpec UserSpec) (*UserSpec, error) {
	err := validateOrGenerateUserIdInSpec(&userSpec)
	if err != nil {
		return nil, err
	}

	var newUser Model
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := svc.objectSvc.Create(txCtx, *userSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
			default:
				return err
			}
		}

		_, err = svc.repo.GetByUserId(txCtx, userSpec.UserId)
		if err == nil {
			return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
		}

		newUserId, err := svc.repo.Create(txCtx, userSpec.ToUser(createdObject.ID))
		if err != nil {
			return err
		}

		newUser, err = svc.repo.GetById(txCtx, newUserId)
		if err != nil {
			return err
		}

		err = svc.eventSvc.TrackResourceCreated(ctx, ResourceTypeUser, newUser.GetUserId(), newUser.ToUserSpec())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newUser.ToUserSpec(), nil
}

func (svc UserService) GetByUserId(ctx context.Context, userId string) (*UserSpec, error) {
	user, err := svc.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user.ToUserSpec(), nil
}

func (svc UserService) List(ctx context.Context, listParams middleware.ListParams) ([]UserSpec, error) {
	userSpecs := make([]UserSpec, 0)

	users, err := svc.repo.List(ctx, listParams)
	if err != nil {
		return userSpecs, err
	}

	for _, user := range users {
		userSpecs = append(userSpecs, *user.ToUserSpec())
	}

	return userSpecs, nil
}

func (svc UserService) UpdateByUserId(ctx context.Context, userId string, userSpec UpdateUserSpec) (*UserSpec, error) {
	currentUser, err := svc.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	currentUser.SetEmail(userSpec.Email)
	err = svc.repo.UpdateByUserId(ctx, userId, currentUser)
	if err != nil {
		return nil, err
	}

	updatedUser, err := svc.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	updatedUserSpec := updatedUser.ToUserSpec()
	err = svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypeUser, updatedUser.GetUserId(), updatedUserSpec)
	if err != nil {
		return nil, err
	}

	return updatedUserSpec, nil
}

func (svc UserService) DeleteByUserId(ctx context.Context, userId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.repo.DeleteByUserId(txCtx, userId)
		if err != nil {
			return err
		}

		err = svc.objectSvc.DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeUser, userId)
		if err != nil {
			return err
		}

		err = svc.eventSvc.TrackResourceDeleted(ctx, ResourceTypeUser, userId, nil)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func validateOrGenerateUserIdInSpec(userSpec *UserSpec) error {
	userIdRegExp := regexp.MustCompile(`^[a-zA-Z0-9_\-\.@\|]+$`)
	if userSpec.UserId != "" {
		// Validate userId if provided
		if !userIdRegExp.Match([]byte(userSpec.UserId)) {
			return service.NewInvalidParameterError("userId", "must be provided and can only contain alphanumeric characters and/or '-', '_', '@', and '|'")
		}
	} else {
		// Generate a UserID for the user if one isn't supplied
		generatedUUID, err := uuid.NewRandom()
		if err != nil {
			return service.NewInternalError("unable to generate random UUID for user")
		}
		userSpec.UserId = generatedUUID.String()
	}
	return nil
}
