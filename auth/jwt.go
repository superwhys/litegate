package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/superwhys/litegate/config"
	"github.com/superwhys/litegate/utils"
)

type jwtAuthenticator struct {
	config *config.Auth
}

func (j *jwtAuthenticator) getValueFromRequest(r *http.Request, claimKey string) string {
	place, name := utils.ParsePlace(claimKey)
	switch place {
	case PlaceHeader:
		return r.Header.Get(name)
	case PlaceQuery:
		return r.URL.Query().Get(name)
	}
	return ""
}

func (j *jwtAuthenticator) Parse(r *http.Request) (Claims, error) {
	token := j.getValueFromRequest(r, j.config.Source)
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	tok, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(j.config.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	result := make(Claims, len(claims))
	for k, v := range claims {
		if v == nil {
			continue
		}
		result[k] = toString(v)
	}
	return result, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
