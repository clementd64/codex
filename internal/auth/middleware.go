package auth

import (
	"context"
	"strings"
	"uuid"

	"connectrpc.com/connect/v2"
	"github.com/clementd64/codex/internal/auth/proto/authconnect"
	"github.com/golang-jwt/jwt/v5"
)

type subjectContextKey struct{}

func SubjectFromContext(ctx context.Context) (uuid.UUID, error) {
	subject, ok := ctx.Value(subjectContextKey{}).(string)
	if !ok {
		return uuid.UUID{}, connect.NewError(connect.CodeUnauthenticated, "authentication required")
	}
	id, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, connect.NewError(connect.CodeUnauthenticated, "invalid token")
	}
	return id, nil
}

func Middleware(signingKey []byte) connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			if spec.Procedure == authconnect.AuthServiceTokenProcedure {
				return next(ctx, spec, stream)
			}

			info, ok := connect.CallInfoForServerContext(ctx)
			if !ok {
				return connect.NewError(connect.CodeUnauthenticated, "bearer token required")
			}

			headers := info.RequestHeader().Values("Authorization")
			if len(headers) != 1 {
				return connect.NewError(connect.CodeUnauthenticated, "bearer token required")
			}
			parts := strings.Fields(headers[0])
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return connect.NewError(connect.CodeUnauthenticated, "bearer token required")
			}

			claims := new(jwt.RegisteredClaims)
			token, err := jwt.ParseWithClaims(
				parts[1],
				claims,
				func(*jwt.Token) (any, error) {
					return signingKey, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithIssuer("codex"),
				jwt.WithAudience("codex"),
				jwt.WithExpirationRequired(),
			)
			if err != nil || !token.Valid || claims.Subject == "" {
				return connect.NewError(connect.CodeUnauthenticated, "invalid bearer token")
			}

			ctx = context.WithValue(ctx, subjectContextKey{}, claims.Subject)
			return next(ctx, spec, stream)
		}
	}
}
