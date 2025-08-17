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
)

var (
	port = flags.Int("port", 8080, "the service listen port")
)

func main() {
	flags.Parse()

	gatewayApp := api.SetupGatewayApp()

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithHttpHandler("/", gatewayApp),
	)
	logging.PanicError(cores.Start(srv, port()))
}
