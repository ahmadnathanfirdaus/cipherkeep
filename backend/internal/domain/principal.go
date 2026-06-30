package domain

// PrincipalKind distinguishes the two kinds of authenticated callers.
type PrincipalKind int

const (
	// PrincipalUser is a human user authenticated via a JWT access token.
	PrincipalUser PrincipalKind = iota
	// PrincipalToken is a service token (API key) authenticated via a bearer key.
	PrincipalToken
)

// Principal is the authenticated caller of a request: either a user or a service
// token. Handlers and services use it to authorize access.
type Principal struct {
	Kind  PrincipalKind
	User  *User
	Token *ServiceToken
}

// IsUser reports whether the principal is a human user.
func (p *Principal) IsUser() bool { return p != nil && p.Kind == PrincipalUser }

// IsToken reports whether the principal is a service token.
func (p *Principal) IsToken() bool { return p != nil && p.Kind == PrincipalToken }
