// File:		server.go
// Created by:	Hoven
// Created on:	2025-08-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package config

import "time"

type GatewayConfig struct {
	// Services list which allowed to be accessed
	Services []string      `json:"services"`
	Timeout  time.Duration `json:"timeout"`
}
