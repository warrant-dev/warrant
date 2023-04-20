package service

import "net/http"

type Route interface {
	GetPattern() string
	GetMethod() string
	GetHandler() http.Handler
	GetOverrideAuthMiddlewareFunc() AuthMiddlewareFunc
	GetDisableAuth() bool
}

type WarrantRoute struct {
	Pattern                    string
	Method                     string
	Handler                    http.Handler
	OverrideAuthMiddlewareFunc AuthMiddlewareFunc
	DisableAuth                bool
	EnableSessionAuth          bool
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

func (route WarrantRoute) GetDisableAuth() bool {
	return route.DisableAuth
}

func (route WarrantRoute) GetEnableSessionAuth() bool {
	return route.EnableSessionAuth
}
