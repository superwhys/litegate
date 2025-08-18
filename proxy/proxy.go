// File:		config.go
// Created by:	Hoven
// Created on:	2025-08-15
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	config "github.com/superwhys/litegate/config"
)

type Agent interface {
	http.Handler
	Auth(w http.ResponseWriter, r *http.Request)
}

type agent struct {
	proxy       *httputil.ReverseProxy
	upstreamURL *url.URL
	auth        *config.Auth
	timeout     time.Duration
}

func NewAgent(auth *config.Auth, upstreamURL string, targetPath string, timeout time.Duration) (*agent, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	if target.Scheme == "" {
		target.Scheme = "http"
	}

	target.Path = targetPath
	proxy := httputil.NewSingleHostReverseProxy(target)

	return &agent{
		proxy:       proxy,
		upstreamURL: target,
		auth:        auth,
		timeout:     timeout,
	}, nil
}

func parsePlace(place string) (string, string) {
	placeSplit := strings.SplitN(place, ".", 2)
	if len(placeSplit) == 1 {
		return "", placeSplit[0]
	}
	return placeSplit[0], placeSplit[1]
}

func (a *agent) getValueFromRequest(r *http.Request, claimKey string) string {
	place, name := parsePlace(claimKey)
	switch place {
	case PlaceHeader:
		return r.Header.Get(name)
	case PlaceQuery:
		return r.URL.Query().Get(name)
	}
	return ""
}

func (a *agent) setValueToRequest(r *http.Request, claimKey string, value string) {
	place, name := parsePlace(claimKey)
	switch place {
	case PlaceHeader:
		r.Header.Set(name, value)
	case PlaceQuery:
		q := r.URL.Query()
		q.Set(name, value)
		r.URL.RawQuery = q.Encode()
	}
}

func (a *agent) Auth(w http.ResponseWriter, r *http.Request) {
	if a.auth == nil {
		return
	}

	auth := a.auth
	token := a.getValueFromRequest(r, auth.Source)
	if token == "" {
		http.Error(w, "token is empty", http.StatusUnauthorized)
		return
	}

	tokenClaims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.Secret), nil
	})
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims := tokenClaims.Claims.(jwt.MapClaims)
	for claimKey, claimName := range auth.Claims {
		value := claims[claimName]
		if value == nil {
			continue
		}

		r = r.WithContext(context.WithValue(r.Context(), claimContextKey(claimKey), value))
	}
}

func (a *agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.auth != nil {
		a.injectAuthData(r)
	}

	timeout := a.timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	r = r.WithContext(ctx)

	a.proxy.ServeHTTP(w, r)
}

func (a *agent) injectAuthData(r *http.Request) {
	for claimKey := range a.auth.Claims {
		value := r.Context().Value(claimContextKey(claimKey))
		if value == nil {
			continue
		}
		valueStr, ok := value.(string)
		if !ok {
			continue
		}

		a.setValueToRequest(r, claimKey, valueStr)
	}
}
