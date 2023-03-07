package authz

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type UserService struct {
	service.BaseService
}

func NewService(env service.Env) UserService {
	return UserService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc UserService) Create(ctx context.Context, userSpec UserSpec) (*UserSpec, error) {
	err := validateOrGenerateUserIdInSpec(&userSpec)
	if err != nil {
		return nil, err
	}

	var newUser *User
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdObject, err := object.NewService(svc.Env()).Create(txCtx, *userSpec.ToObjectSpec())
		if err != nil {
			switch err.(type) {
			case *service.DuplicateRecordError:
				return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
			default:
				return err
			}
		}

		userRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		_, err = userRepository.GetByUserId(txCtx, userSpec.UserId)
		if err == nil {
			return service.NewDuplicateRecordError("User", userSpec.UserId, "A user with the given userId already exists")
		}

		newUserId, err := userRepository.Create(txCtx, *userSpec.ToUser(createdObject.ID))
		if err != nil {
			return err
		}

		newUser, err = userRepository.GetById(txCtx, newUserId)
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
	userRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	user, err := userRepository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user.ToUserSpec(), nil
}

func (svc UserService) List(ctx context.Context, listParams middleware.ListParams) ([]UserSpec, error) {
	userSpecs := make([]UserSpec, 0)
	userRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	users, err := userRepository.List(ctx, listParams)
	if err != nil {
		return userSpecs, err
	}

	for _, user := range users {
		userSpecs = append(userSpecs, *user.ToUserSpec())
	}

	return userSpecs, nil
}

func (svc UserService) UpdateByUserId(ctx context.Context, userId string, userSpec UpdateUserSpec) (*UserSpec, error) {
	userRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentUser, err := userRepository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	currentUser.Email = userSpec.Email
	err = userRepository.UpdateByUserId(ctx, userId, *currentUser)
	if err != nil {
		return nil, err
	}

	updatedUser, err := userRepository.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	return updatedUser.ToUserSpec(), nil
}

func (svc UserService) DeleteByUserId(ctx context.Context, userId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		userRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = userRepository.DeleteByUserId(txCtx, userId)
		if err != nil {
			return err
		}

		err = object.NewService(svc.Env()).DeleteByObjectTypeAndId(txCtx, objecttype.ObjectTypeUser, userId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func validateOrGenerateUserIdInSpec(userSpec *UserSpec) error {
	userIdRegExp := regexp.MustCompile(`^[a-zA-Z0-9_\-\.@]+$`)
	if userSpec.UserId != "" {
		// Validate userId if provided
		if !userIdRegExp.Match([]byte(userSpec.UserId)) {
			return service.NewInvalidParameterError("userId", "must be provided and can only contain alphanumeric characters and/or '-', '_', and '@'")
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
