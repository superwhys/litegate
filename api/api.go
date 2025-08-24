// File:		api.go
// Created by:	Hoven
// Created on:	2025-08-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package api

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/litegate/api/middleware"
	"github.com/superwhys/litegate/api/router"
	"github.com/superwhys/litegate/config"
)

func SetupGatewayApp(configLoader config.ProxyConfigLoader) http.Handler {
	app := ginutils.NewServerHandler(
		// debug group
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/debug"),
			ginutils.WithRouterHandler(router.DebugRouter(configLoader)),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/__:serviceName"),
			ginutils.WithMiddleware(middleware.ParseProxyConfig(configLoader)),
			ginutils.WithAnyHandler("/*any", router.ProxyRouter()),
		),
	)

	return app
}
