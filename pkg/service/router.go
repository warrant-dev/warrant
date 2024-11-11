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

package service

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/stats"
	"github.com/warrant-dev/warrant/pkg/wookie"
)

type RouteHandler[T Service] struct {
	svc     T
	handler func(svc T, w http.ResponseWriter, r *http.Request) error
}

func NewRouteHandler[T Service](svc T, handler func(svc T, w http.ResponseWriter, r *http.Request) error) RouteHandler[T] {
	return RouteHandler[T]{
		svc:     svc,
		handler: handler,
	}
}

func (rh RouteHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := rh.handler(rh.svc, w, r)
	if err != nil {
		// Write err response to client
		SendErrorResponse(w, err)

		// Log and send err to Sentry
		logEvent := hlog.FromRequest(r).Error().Stack().Err(err)
		if apiError, ok := err.(Error); ok {
			// Add additional context to log if ApiError
			logEvent = logEvent.Str("apiError", apiError.GetTag()).
				Int("statusCode", apiError.GetStatus())
		}

		// Log event
		logEvent.Msg("error log")
	}
}

func NewRouter(config config.Config, pathPrefix string, routes []Route, authMiddleware AuthMiddlewareFunc, routerMiddlewares []Middleware, requestMiddlewares []Middleware) (*mux.Router, error) {
	router := mux.NewRouter()

	var logger zerolog.Logger
	logFile, err := os.Create("warrant.log")
	if err != nil {
		logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Logger().
			Level(zerolog.Level(config.GetLogLevel()))
	} else {
		logger = zerolog.New(logFile).
			With().
			Timestamp().
			Logger().
			Level(zerolog.Level(config.GetLogLevel()))
	}
	// Setup default middleware
	//if logger.GetLevel() == zerolog.DebugLevel {
	//	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	//}
	router.Use(hlog.NewHandler(logger))
	router.Use(stats.RequestStatsMiddleware)
	router.Use(wookie.WarrantTokenMiddleware)
	if config.GetEnableAccessLog() {
		router.Use(accessLogMiddleware)
	}
	router.Use(hlog.RequestIDHandler("requestId", "Warrant-Request-Id"))
	router.Use(hlog.URLHandler("uri"))
	router.Use(hlog.MethodHandler("method"))
	router.Use(hlog.ProtoHandler("protocol"))

	// Setup router middlewares, which will be run on ALL
	// requests, even if they are to non-existent endpoints.
	for _, routerMiddleware := range routerMiddlewares {
		router.Use(mux.MiddlewareFunc(routerMiddleware))
	}

	// Setup routes
	for _, route := range routes {
		routePattern := fmt.Sprintf("%s%s", pathPrefix, route.GetPattern())
		middlewareWrappedHandler := ChainMiddleware(route.GetHandler(), requestMiddlewares...)

		var err error
		if route.GetOverrideAuthMiddlewareFunc() != nil {
			middlewareWrappedHandler, err = route.GetOverrideAuthMiddlewareFunc()(config, middlewareWrappedHandler)
		} else {
			middlewareWrappedHandler, err = authMiddleware(config, middlewareWrappedHandler)
		}
		if err != nil {
			return nil, err
		}

		router.Handle(routePattern, middlewareWrappedHandler).Methods(route.GetMethod())
	}

	// Configure catch all handler for 404s
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendErrorResponse(w, NewRecordNotFoundError("Endpoint", r.URL.Path))
	}))

	return router, nil
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		logger := hlog.FromRequest(r)
		logEvent := logger.Info().
			Str("method", r.Method).
			Str("protocol", r.Proto).
			Stringer("uri", r.URL).
			Int("statusCode", status).
			Int("size", size).
			Dur("duration", duration).
			Str("clientIp", GetClientIpAddress(r))

		if referer := r.Referer(); referer != "" {
			logEvent = logEvent.Str("referer", referer)
		}

		if userAgent := r.UserAgent(); userAgent != "" {
			logEvent = logEvent.Str("userAgent", userAgent)
		}

		logEvent.Msg("access log")
	})(next)
}

func GetClientIpAddress(r *http.Request) string {
	clientIpAddress := r.Header.Get("X-Forwarded-For")
	if clientIpAddress == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}

	return clientIpAddress
}
