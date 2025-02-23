package client

import (
	"context"
	"net/http"

	"github.com/waro163/cloud-root/http/client/auth"
	"github.com/waro163/requests/hooks"
)

type Options func(*BaseClient)

func WithAuthSet(authSet auth.IAuthSet) Options {
	return func(bc *BaseClient) {
		bc.Client.Use(hooks.NewCustomHook(
			func(ctx context.Context, req *http.Request) error {
				if err := authSet.SetRequest(req); err != nil {
					authSet.Reset(ctx)
					return err
				}
				return nil
			},
			nil,
			func(ctx context.Context, req *http.Request, resp *http.Response) error {
				if resp != nil && resp.StatusCode == http.StatusUnauthorized {
					authSet.Reset(ctx)
				}
				return nil
			},
			nil),
		)
	}
}
