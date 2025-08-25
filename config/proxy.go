// File:		proxy.go
// Created by:	Hoven
// Created on:	2025-08-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package config

import (
	"context"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/miebyte/goutils/logging"
)

type ProxyConfigLoader interface {
	Get(service string) (*RouteConfig, error)
	GetAll() ([]*RouteConfig, error)
	Watch() error
}

type Upstream struct {
	Auth        *Auth
	Timeout     time.Duration
	UpstreamURL string
	TargetPath  string
}

// RouteConfig 路由配置
type RouteConfig struct {
	// 代理地址列表
	Proxy ProxyConfig `json:"proxy" validate:"required"`
	// 超时时间（选填，默认30秒）
	Timeout string `json:"timeout"`
	// 身份验证配置（选填）
	Auth *Auth `json:"auth,omitempty"`
	// 路由配置（必填）
	Routes []Route `json:"routes" validate:"required,min=1"`
}

func (rc *RouteConfig) MatchRequest(ctx context.Context, req *http.Request) *Upstream {
	for _, route := range rc.Routes {
		if route.Match == "" {
			continue
		}

		logging.Debugc(ctx, "try match route: %s", logging.JsonifyNoIndent(route))
		regex, err := regexp.Compile(route.Match)
		if err != nil {
			logging.Errorf("compile regex error: %v", err)
			return nil
		}

		timeoutStr := rc.Timeout
		if route.Timeout != "" {
			timeoutStr = route.Timeout
		}

		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			logging.Errorf("parse timeout error: %v", err)
			return nil
		}

		auth := rc.Auth
		if route.DisableAuth {
			auth = nil
		} else if route.Auth != nil {
			auth = route.Auth
		}

		if regex.MatchString(req.URL.Path) {
			upstream := &Upstream{
				Auth:        auth,
				Timeout:     timeout,
				UpstreamURL: route.Proxy.pickAddress(),
				TargetPath:  req.URL.Path,
			}
			logging.Debugc(ctx, "matched route: %s", logging.JsonifyNoIndent(upstream))
			return upstream
		}
	}

	return nil
}

// Auth 身份验证配置
type Auth struct {
	// Token类型，固定为jwt
	Type string `json:"type" validate:"required,eq=jwt"`
	// Token在请求中的位置
	// $header.token
	// $query.token
	Source string `json:"source" validate:"required"`
	// JWT密钥
	Secret string `json:"secret" validate:"required"`
	// JWT解码后数据存储位置映射
	// {
	// 	"$query.user_id": "user_id",
	// 	"$header.X-User": "userName",
	// }
	// 该示例表示
	// 1. 将JWT解码后的 `user_id` 数据存储到请求 query中, key为 user_id
	// 2. 将JWT解码后的 `userName` 数据存储到请求 header中, key为 X-User
	Claims map[string]string `json:"claims" validate:"required"`
}

// Route 路由配置
type Route struct {
	// URL匹配 (正则表达式)
	Match string `json:"match" validate:"required"`
	// 代理地址列表
	Proxy ProxyConfig `json:"proxy"`
	// 超时时间（选填）
	Timeout string `json:"timeout"`
	// 是否禁用身份验证
	DisableAuth bool `json:"disable_auth"`
	// 身份验证配置覆盖
	Auth *Auth `json:"auth,omitempty"`
}

type ProxyConfig []string

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (p ProxyConfig) pickAddress() string {
	if len(p) == 0 {
		return ""
	}
	if len(p) == 1 {
		return p[0]
	}
	return p[rand.Intn(len(p))]
}
