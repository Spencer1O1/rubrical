package auth

import (
	"context"
	"errors"
)

type contextKey int

const userContextKey contextKey = iota

var ErrNoUser = errors.New("not authenticated")

type User struct {
	ID          int64
	Email       string
	DisplayName string
}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (User, error) {
	user, ok := ctx.Value(userContextKey).(User)
	if !ok || user.ID <= 0 {
		return User{}, ErrNoUser
	}
	return user, nil
}

func UserID(ctx context.Context) (int64, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}
