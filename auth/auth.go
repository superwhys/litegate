package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/superwhys/litegate/config"
	"github.com/superwhys/litegate/utils"
)

const (
	AuthTypeJWT = "jwt"

	PlaceHeader = "$header"
	PlaceQuery  = "$query"
)

type (
	ClaimContextKey string
	Claims          map[string]string
)

type Authenticator interface {
	Parse(r *http.Request) (Claims, error)
}

func NewAuthenticator(cfg *config.Auth) (Authenticator, error) {
	if cfg == nil {
		return nil, nil
	}
	switch cfg.Type {
	case AuthTypeJWT:
		return &jwtAuthenticator{cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Type)
	}
}

func InjectClaimsToContext(r *http.Request, claims Claims) context.Context {
	ctx := r.Context()
	for place, claimKey := range claims {
		value, ok := claims[claimKey]
		if !ok || value == "" {
			continue
		}
		ctx = context.WithValue(ctx, ClaimContextKey(place), value)
	}
	return ctx
}

func setValueToRequest(r *http.Request, claimKey string, value string) {
	place, name := utils.ParsePlace(claimKey)
	switch place {
	case PlaceHeader:
		r.Header.Set(name, value)
	case PlaceQuery:
		q := r.URL.Query()
		q.Set(name, value)
		r.URL.RawQuery = q.Encode()
	}
}

func InjectClaimsToRequest(r *http.Request, auth *config.Auth) {
	for place := range auth.Claims {
		value := r.Context().Value(ClaimContextKey(place))
		if value == nil {
			continue
		}
		valueStr, ok := value.(string)
		if !ok {
			continue
		}

		setValueToRequest(r, place, valueStr)
	}
}
