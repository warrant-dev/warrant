package service

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/hlog"
	"github.com/warrant-dev/warrant/pkg/config"
)

const FirebasePublicKeyUrl = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

type key int

const (
	authInfoKey key = iota
)

const (
	EnableSessionAuthKey = "EnableSessionAuth"
)

type AuthInfo struct {
	UserId   string
	TenantId string
}

type AuthMiddlewareFunc func(next http.Handler, config *config.Config, args map[string]interface{}) http.Handler

func DefaultAuthMiddleware(next http.Handler, config *config.Config, args map[string]interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			SendErrorResponse(w, NewUnauthorizedError("Request missing Authorization header"))
			return
		}

		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 {
			SendErrorResponse(w, NewUnauthorizedError("Invalid Authorization header"))
			logger.Warn().Msgf("Invalid Authorization header %s", authHeader)
			return
		}

		tokenType := authHeaderParts[0]
		tokenString := authHeaderParts[1]

		var authInfo *AuthInfo
		switch tokenType {
		case "ApiKey":
			if !secureCompareEqual(tokenString, config.ApiKey) {
				SendErrorResponse(w, NewUnauthorizedError("Invalid API key"))
				return
			}
			authInfo = &AuthInfo{}
		case "Bearer":
			if enableSessionAuth, ok := args["enableSessionAuth"].(bool); ok {
				if !enableSessionAuth {
					SendErrorResponse(w, NewUnauthorizedError("Error validating token"))
					logger.Err(fmt.Errorf("invalid authentication for the endpoint")).Msg("Session authentication not supported for this endpoint")
					return
				}
			} else {
				SendErrorResponse(w, NewUnauthorizedError("Error validating token"))
				logger.Err(fmt.Errorf("enableSessionAuth must be of type bool"))
				return
			}

			if config.Authentication.Provider == "" {
				SendErrorResponse(w, NewInternalError("Error validating token"))
				logger.Err(fmt.Errorf("invalid authentication provider configuration")).Msg("Must configure an authentication provider to allow requests that use third party auth tokens.")
				return
			}

			var publicKey *rsa.PublicKey
			var publicKeys map[string]string
			var err error
			switch config.Authentication.Provider {
			case "firebase":
				// Retrieve Firebase public keys
				response, err := http.Get(FirebasePublicKeyUrl)
				if err != nil {
					SendErrorResponse(w, NewInternalError("Error validating token"))
					logger.Err(err).Msg("Error fetching Firebase public keys")
					return
				}

				defer response.Body.Close()

				contents, err := io.ReadAll(response.Body)
				if err != nil {
					SendErrorResponse(w, NewInternalError("Error validating token"))
					logger.Err(err).Msg("Error reading Firebase public keys")
					return
				}

				json.Unmarshal(contents, &publicKeys)
			default:
				publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(config.Authentication.PublicKey))
				if err != nil {
					SendErrorResponse(w, NewInternalError("Error validating token"))
					logger.Err(fmt.Errorf("invalid authentication provider configuration")).Msg("Invalid public key for configured authentication provider")
					return
				}
			}

			checkedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, NewUnauthorizedError(fmt.Sprintf("Invalid %s token: unexpected signing method %v", tokenType, token.Header["alg"]))
				}

				if config.Authentication.Provider == "firebase" {
					publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeys[token.Header["kid"].(string)]))
					if err != nil {
						return nil, NewUnauthorizedError("Invalid token")
					}
				}

				return publicKey, nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					SendErrorResponse(w, NewTokenExpiredError())
					return
				}

				SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
				return
			}

			if !checkedToken.Valid {
				SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
				return
			}

			// Get claims
			tokenClaims := checkedToken.Claims.(jwt.MapClaims)

			if _, ok := tokenClaims[config.Authentication.UserIdClaim]; !ok {
				SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
				logger.Warn().Msgf("Unable to retrieve user id from token with given identifier: %s", config.Authentication.UserIdClaim)
				return
			}
			userId := tokenClaims[config.Authentication.UserIdClaim].(string)

			authInfo = &AuthInfo{
				UserId: userId,
			}

			if config.Authentication.TenantIdClaim != "" {
				if _, ok := tokenClaims[config.Authentication.TenantIdClaim]; !ok {
					SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
					logger.Warn().Msgf("Unable to retrieve tenant id from token with given identifier: %s", config.Authentication.TenantIdClaim)
				}
				authInfo.TenantId = tokenClaims[config.Authentication.TenantIdClaim].(string)
			}
		default:
			SendErrorResponse(w, NewUnauthorizedError("Invalid Authorization header prefix. Must be ApiKey or Bearer"))
			return
		}

		// Add authInfo to request context
		newContext := context.WithValue(r.Context(), authInfoKey, *authInfo)

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}

// GetAuthInfoFromRequestContext returns the AuthInfo object from the given context
func GetAuthInfoFromRequestContext(context context.Context) *AuthInfo {
	contextVal := context.Value(authInfoKey)
	if contextVal != nil {
		authInfo := context.Value(authInfoKey).(AuthInfo)
		return &authInfo
	}

	return nil
}

func secureCompareEqual(given string, actual string) bool {
	if subtle.ConstantTimeEq(int32(len(given)), int32(len(actual))) == 1 {
		return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
	} else {
		return subtle.ConstantTimeCompare([]byte(actual), []byte(actual)) == 1 && false
	}
}
