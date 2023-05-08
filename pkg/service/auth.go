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
	AuthTypeApiKey = "ApiKey"
	AuthTypeBearer = "Bearer"
)

type AuthInfo struct {
	UserId   string
	TenantId string
}

type AuthMiddlewareFunc func(config config.Config, next http.Handler) (http.Handler, error)

func ApiKeyAuthMiddleware(cfg config.Config, next http.Handler) (http.Handler, error) {
	warrantCfg, ok := cfg.(config.WarrantConfig)
	if !ok {
		return nil, fmt.Errorf("cfg parameter on DefaultAuthMiddleware must be a WarrantConfig")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, tokenString, err := parseAuthTokenFromRequest(r, []string{AuthTypeApiKey})
		if err != nil {
			SendErrorResponse(w, NewUnauthorizedError(fmt.Sprintf("Invalid authorization header: %s", err.Error())))
			return
		}

		if !secureCompareEqual(tokenString, warrantCfg.GetAuthentication().ApiKey) {
			SendErrorResponse(w, NewUnauthorizedError("Invalid API key"))
			return
		}

		newContext := context.WithValue(r.Context(), authInfoKey, &AuthInfo{})
		next.ServeHTTP(w, r.WithContext(newContext))
	}), nil
}

func ApiKeyAndSessionAuthMiddleware(cfg config.Config, next http.Handler) (http.Handler, error) {
	warrantCfg, ok := cfg.(config.WarrantConfig)
	if !ok {
		return nil, fmt.Errorf("cfg parameter on DefaultAuthMiddleware must be a WarrantConfig")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		tokenType, tokenString, err := parseAuthTokenFromRequest(r, []string{AuthTypeApiKey, AuthTypeBearer})
		if err != nil {
			SendErrorResponse(w, NewUnauthorizedError(fmt.Sprintf("Invalid authorization header: %s", err.Error())))
			return
		}

		var authInfo *AuthInfo
		switch tokenType {
		case AuthTypeApiKey:
			if !secureCompareEqual(tokenString, warrantCfg.GetAuthentication().ApiKey) {
				SendErrorResponse(w, NewUnauthorizedError("Invalid API key"))
				return
			}

			authInfo = &AuthInfo{}
		case AuthTypeBearer:
			if warrantCfg.GetAuthentication().Provider == nil {
				SendErrorResponse(w, NewInternalError("Error validating token"))
				logger.Err(fmt.Errorf("invalid authentication provider configuration")).Msg("Must configure an authentication provider to allow requests that use third party auth tokens.")
				return
			}

			var publicKey *rsa.PublicKey
			var publicKeys map[string]string
			var err error
			switch warrantCfg.GetAuthentication().Provider.Name {
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
				publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(warrantCfg.GetAuthentication().Provider.PublicKey))
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

				if warrantCfg.GetAuthentication().Provider.Name == "firebase" {
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
			if _, ok := tokenClaims[warrantCfg.GetAuthentication().Provider.UserIdClaim]; !ok {
				SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
				logger.Warn().Msgf("Unable to retrieve user id from token with given identifier: %s", warrantCfg.GetAuthentication().Provider.UserIdClaim)
				return
			}

			authInfo = &AuthInfo{
				UserId: tokenClaims[warrantCfg.GetAuthentication().Provider.UserIdClaim].(string),
			}

			if warrantCfg.GetAuthentication().Provider.TenantIdClaim != "" {
				if _, ok := tokenClaims[warrantCfg.GetAuthentication().Provider.TenantIdClaim]; !ok {
					SendErrorResponse(w, NewUnauthorizedError("Invalid token"))
					logger.Warn().Msgf("Unable to retrieve tenant id from token with given identifier: %s", warrantCfg.GetAuthentication().Provider.TenantIdClaim)
				}
				authInfo.TenantId = tokenClaims[warrantCfg.GetAuthentication().Provider.TenantIdClaim].(string)
			}
		}

		newContext := context.WithValue(r.Context(), authInfoKey, *authInfo)
		next.ServeHTTP(w, r.WithContext(newContext))
	}), nil
}

func PassthroughAuthMiddleware(cfg config.Config, next http.Handler) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}), nil
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

func parseAuthTokenFromRequest(r *http.Request, validTokenTypes []string) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 {
		return "", "", fmt.Errorf("invalid format")
	}

	authTokenType := authHeaderParts[0]
	authToken := authHeaderParts[1]

	var isValidTokenType bool
	for _, validTokenType := range validTokenTypes {
		if authTokenType == validTokenType {
			isValidTokenType = true
		}
	}
	if !isValidTokenType {
		return "", "", fmt.Errorf("authorization header prefix must be one of: %s", strings.Join(validTokenTypes, ", "))
	}

	return authTokenType, authToken, nil
}

func secureCompareEqual(given string, actual string) bool {
	if subtle.ConstantTimeEq(int32(len(given)), int32(len(actual))) == 1 {
		return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
	} else {
		return subtle.ConstantTimeCompare([]byte(actual), []byte(actual)) == 1 && false
	}
}
