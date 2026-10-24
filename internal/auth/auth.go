package auth

import (
	"context"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/clementd64/codex/internal/app"
	authpb "github.com/clementd64/codex/internal/auth/proto"
	"github.com/clementd64/codex/internal/auth/proto/authconnect"
	authquery "github.com/clementd64/codex/internal/auth/query"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

type AuthServer struct {
	app   *app.App
	query *authquery.Queries
}

func (s *AuthServer) Token(
	ctx context.Context,
	req *authpb.TokenRequest,
) (*authpb.TokenResponse, error) {
	provider, err := oidc.NewProvider(ctx, s.app.OIDCDiscoveryEndpoint)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "OIDC provider discovery failed")
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: s.app.OIDCClientID})
	idToken, err := verifier.Verify(ctx, req.GetToken())
	if err != nil || idToken.Subject == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, "invalid ID token")
	}

	id, err := s.query.UpsertUser(ctx, idToken.Subject)
	if err != nil {
		return nil, err
	}

	// cap max expiration to 1h
	now := time.Now()
	expires := now.Add(time.Hour)
	if idToken.Expiry.Before(expires) {
		expires = idToken.Expiry
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "codex",
		Subject:   id.String(),
		Audience:  jwt.ClaimStrings{"codex"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
	}).SignedString(s.app.SigningKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, "could not sign JWT")
	}
	return &authpb.TokenResponse{Token: signed, Expiration: uint64(expires.Unix())}, nil
}

func Register(server *connect.Server, application *app.App) {
	authconnect.RegisterAuthServiceHandler(server, &AuthServer{
		app:   application,
		query: authquery.New(application.DB),
	})
}
