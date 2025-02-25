package auth

import (
	"context"
	"net/http"
)

type IAuthSet interface {
	SetRequest(req *http.Request) error
	Reset(ctx context.Context)
}

var (
	_ IAuthSet = (*m2mAuthSet)(nil)
)
