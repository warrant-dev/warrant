package service

import "net/http"

type Route interface {
	GetPattern() string
	GetMethod() string
	GetHandler() http.Handler
	GetOverrideAuthMiddlewareFunc() AuthMiddlewareFunc
}

type WarrantRoute struct {
	Pattern                    string
	Method                     string
	Handler                    http.Handler
	OverrideAuthMiddlewareFunc AuthMiddlewareFunc
}

func (route WarrantRoute) GetPattern() string {
	return route.Pattern
}

func (route WarrantRoute) GetMethod() string {
	return route.Method
}

func (route WarrantRoute) GetHandler() http.Handler {
	return route.Handler
}

func (route WarrantRoute) GetOverrideAuthMiddlewareFunc() AuthMiddlewareFunc {
	return route.OverrideAuthMiddlewareFunc
}
