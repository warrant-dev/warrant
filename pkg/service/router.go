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
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/config"
)

type Route struct {
	Pattern           string
	Method            string
	Handler           http.Handler
	AuthMiddleware    AuthMiddlewareFunc
	DisableAuth       bool
	EnableSessionAuth bool
}

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
		logEvent.Msg("ERROR")
	}
}

func NewRouter(config config.Config, pathPrefix string, routes []Route, authMiddleware AuthMiddlewareFunc, additionalMiddlewares ...mux.MiddlewareFunc) *mux.Router {
	router := mux.NewRouter()

	// Setup default middleware
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Level(zerolog.Level(config.GetLogLevel()))
	if logger.GetLevel() == zerolog.DebugLevel {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	router.Use(hlog.NewHandler(logger))
	if config.GetEnableAccessLog() {
		router.Use(accessLogMiddleware)
		router.Use(hlog.RequestIDHandler("requestId", "Warrant-Request-Id"))
	}

	router.Use(hlog.URLHandler("uri"))

	if authMiddleware == nil {
		authMiddleware = DefaultAuthMiddleware
	}

	// Setup supplied middleware
	for _, additionalMiddleware := range additionalMiddlewares {
		router.Use(additionalMiddleware)
	}

	// Setup routes
	for _, route := range routes {
		log.Debug().Msgf("Route: %+v Auth middleware? %+v", route.Pattern, route.AuthMiddleware)
		defaultOptions := map[string]interface{}{
			EnableSessionAuthKey: route.EnableSessionAuth,
		}
		routePattern := fmt.Sprintf("%s%s", pathPrefix, route.Pattern)
		if route.DisableAuth || config.GetApiKey() == "" {
			router.Handle(routePattern, route.Handler).Methods(route.Method)
		} else if route.AuthMiddleware != nil {
			router.Handle(routePattern, route.AuthMiddleware(route.Handler, config, defaultOptions)).Methods(route.Method)
		} else {
			router.Handle(routePattern, authMiddleware(route.Handler, config, defaultOptions)).Methods(route.Method)
		}
	}

	// Configure catch all handler for 404s
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendErrorResponse(w, NewRecordNotFoundError("Endpoint", r.URL.Path))
	}))

	return router
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		logEvent := hlog.FromRequest(r).Info().
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

		if duration.Milliseconds() >= 500 {
			logEvent = logEvent.Bool("slow", true)
		}

		logEvent.Msg("ACCESS")
	})(next)
}

func GetClientIpAddress(r *http.Request) string {
	clientIpAddress := r.Header.Get("X-Forwarded-For")
	if clientIpAddress == "" {
		return strings.Split(r.RemoteAddr, ":")[0]
	}

	return clientIpAddress
}
