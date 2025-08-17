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
)

func SetupGatewayApp() http.Handler {
	app := ginutils.NewServerHandler(
		// debug group
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("debug"),
			ginutils.WithRouterHandler(),
		),
	)

	return app
}
