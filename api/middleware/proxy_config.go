// File:		proxy_config.go
// Created by:	Hoven
// Created on:	2025-08-22
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/litegate/config"
)

const (
	ProxyConfigKey = "proxyConfig"
)

func GetProxyConfig(c *gin.Context) *config.RouteConfig {
	route, ok := c.Get(ProxyConfigKey)
	if !ok {
		return nil
	}
	return route.(*config.RouteConfig)
}

func ParseProxyConfig(gatewayConf *config.GatewayConfig, configLoader config.ProxyConfigLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		if !slices.Contains(gatewayConf.Services, serviceName) {
			ginutils.ReturnError(c, http.StatusOK, "service not found")
			return
		}

		route, err := configLoader.Get(serviceName)
		if err != nil {
			ginutils.ReturnError(c, http.StatusOK, err.Error())
			return
		}

		logging.Debugc(c, "proxy route config: %+v", route)

		targetPath := strings.TrimPrefix(c.Request.URL.Path, "/__"+serviceName)

		ctx := logging.With(c.Request.Context(), "proxyService", serviceName)
		ctx = logging.With(ctx, "proxyTargetPath", targetPath)

		c.Request = c.Request.WithContext(ctx)
		c.Request.URL.Path = targetPath

		c.Set(ProxyConfigKey, route)
		c.Next()
	}
}
