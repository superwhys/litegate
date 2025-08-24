// File:		main.go
// Created by:	Hoven
// Created on:	2025-08-15
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/litegate/api"
	"github.com/superwhys/litegate/config"
	"github.com/superwhys/litegate/config/loader"
)

var (
	port       = flags.Int("port", 8080, "the service listen port")
	configFlag = flags.Struct("config", (*config.GatewayConfig)(nil), "gateway server config")
)

func main() {
	flags.Parse()

	gatewayConfig := new(config.GatewayConfig)
	logging.PanicError(configFlag(gatewayConfig))

	proxyConfigLoader := loader.NewLocalConfigLoader("./content/proxy")
	gatewayApp := api.SetupGatewayApp(proxyConfigLoader)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithHttpHandler("/", gatewayApp),
	)
	logging.PanicError(cores.Start(srv, port()))
}
