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

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

func SetupGatewayApp() http.Handler {
	app := ginutils.NewServerHandler(
		// debug group
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("debug"),
			ginutils.WithRouterHandler(),
		),
		ginutils.WithAnyHandler("/__:serviceName/*any", func(c *gin.Context) {
			serviceName := c.Param("serviceName")
			path := c.Param("any")

			// 验证服务名称格式
			if serviceName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
				return
			}

			// 这里可以根据 serviceName 进行不同的处理逻辑
			c.JSON(http.StatusOK, gin.H{
				"message": "Service Gateway",
				"service": serviceName,
				"path":    path,
			})
		}),
	)

	return app
}
