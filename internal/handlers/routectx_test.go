package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
)

func routeCtx(params map[string]string) *chi.Context {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return rctx
}

func withRouteCtx(ctx context.Context, rctx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}
