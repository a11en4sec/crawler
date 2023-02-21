package auth

import (
	"context"
	"errors"

	"go-micro.dev/v4"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
)

func NewAuthWrapper(service micro.Service) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// 登录入口不验证token
			if req.Method() == "User.UserLogin" {
				return h(ctx, req, rsp)
			}

			// Fetch metadata from context (request headers).
			md, b := metadata.FromContext(ctx)
			if !b {
				return errors.New("no metadata found")
			}

			// Get auth header.
			//authHeader, ok := md["Authorization"]
			//if !ok || !strings.HasPrefix(authHeader, auth.BearerScheme) {
			//	return errors.New("no auth token provided")
			//}

			// Extract auth token.
			//token := strings.TrimPrefix(authHeader, auth.BearerScheme)

			token, ok := md["Authorization"]
			if !ok {
				return errors.New("no auth token provided")
			}

			// Extract account from token.
			a := service.Options().Auth
			_, err := a.Inspect(token)
			if err != nil {
				return errors.New("auth token invalid")
			}

			return h(ctx, req, rsp)
		}
	}
}
