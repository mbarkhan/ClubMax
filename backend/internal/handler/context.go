package handler

import (
	"context"
	"github.com/clubmax/clubmax/internal/domain"
	"net/http"
)

type contextKey int

const userKey contextKey = 1

func withUser(c context.Context, u *domain.User) context.Context {
	return context.WithValue(c, userKey, u)
}
func current(r *http.Request) *domain.User {
	u, _ := r.Context().Value(userKey).(*domain.User)
	return u
}
