// File:		agent.go
// Created by:	Hoven
// Created on:	2025-08-15
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package agent

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
	auth          *config.Auth
	timeout       time.Duration
	authenticator auth.Authenticator
}

func NewAgent(upstreamConf *config.Upstream) (*agent, error) {
	target, err := url.Parse(upstreamConf.UpstreamURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = upstreamConf.TargetPath
		req.URL.RawPath = upstreamConf.TargetPath

		targetQuery := target.RawQuery
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	authenticator, err := auth.NewAuthenticator(upstreamConf.Auth)
	if err != nil {
		return nil, err
	}

	return &agent{
		proxy:         proxy,
		auth:          upstreamConf.Auth,
		timeout:       upstreamConf.Timeout,
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
