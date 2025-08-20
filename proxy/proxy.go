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
	"time"

	"github.com/superwhys/litegate/auth"
	"github.com/superwhys/litegate/config"
)

type Agent interface {
	http.Handler
	Auth(w http.ResponseWriter, r *http.Request)
}

type agent struct {
	proxy         *httputil.ReverseProxy
	upstreamURL   *url.URL
	auth          *config.Auth
	timeout       time.Duration
	authenticator auth.Authenticator
}

// NewAgent creates a new agent with the given auth, upstream URL, target path, and timeout.
// The agent is a HTTP handler that proxies requests to the upstream URL.
// The agent also injects the auth data into the request.
// The agent also sets the timeout for the request.
func NewAgent(authConfig *config.Auth, upstreamURL string, targetPath string, timeout time.Duration) (*agent, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	if target.Scheme == "" {
		target.Scheme = "http"
	}

	target.Path = targetPath
	proxy := httputil.NewSingleHostReverseProxy(target)
	authenticator, err := auth.NewAuthenticator(authConfig)
	if err != nil {
		return nil, err
	}

	return &agent{
		proxy:         proxy,
		upstreamURL:   target,
		auth:          authConfig,
		timeout:       timeout,
		authenticator: authenticator,
	}, nil
}

func (a *agent) Auth(w http.ResponseWriter, r *http.Request) {
	if a.authenticator == nil {
		return
	}

	claims, err := a.authenticator.Parse(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	*r = *r.WithContext(auth.InjectClaimsToContext(r, claims))
}

func (a *agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.auth != nil {
		auth.InjectClaimsToRequest(r, a.auth)
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
