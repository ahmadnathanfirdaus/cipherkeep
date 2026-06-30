package service

import (
	"context"

	"github.com/cipherkeep/backend/internal/domain"
)

// Authenticator resolves a bearer credential into a principal. A credential with the
// service-token prefix is validated as a service token; otherwise it is parsed as a
// user JWT access token.
type Authenticator struct {
	users  *AuthService
	tokens *TokenService
}

// NewAuthenticator wires the composite authenticator.
func NewAuthenticator(users *AuthService, tokens *TokenService) *Authenticator {
	return &Authenticator{users: users, tokens: tokens}
}

// AuthenticatePrincipal implements middleware.Authenticator.
func (a *Authenticator) AuthenticatePrincipal(ctx context.Context, credential string) (*domain.Principal, error) {
	if HasServiceTokenPrefix(credential) {
		token, err := a.tokens.AuthenticateToken(ctx, credential)
		if err != nil {
			return nil, err
		}
		return &domain.Principal{Kind: domain.PrincipalToken, Token: token}, nil
	}

	user, err := a.users.Authenticate(ctx, credential)
	if err != nil {
		return nil, err
	}
	return &domain.Principal{Kind: domain.PrincipalUser, User: user}, nil
}
