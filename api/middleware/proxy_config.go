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

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
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

func ParseProxyConfig(configLoader config.ProxyConfigLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		route, err := configLoader.Get(serviceName)
		if err != nil {
			ginutils.ReturnError(c, http.StatusOK, err.Error())
			return
		}
		c.Set(ProxyConfigKey, route)
		c.Next()
	}
}
