// File:		transport.go
// Created by:	Hoven
// Created on:	2025-08-26
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.
package agent

import (
	"net"
	"net/http"

	"github.com/superwhys/litegate/config"
)

func customTransport(transportConf *config.TransportConfig) *http.Transport {
	return &http.Transport{
		MaxIdleConns:        transportConf.MaxIdleConns,
		MaxIdleConnsPerHost: transportConf.MaxIdleConnsPerHost,
		MaxConnsPerHost:     transportConf.MaxConnsPerHost,
		IdleConnTimeout:     transportConf.IdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   transportConf.TcpDialTimeout,
			KeepAlive: transportConf.KeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   transportConf.TcpHandshakeTimeout,
		ResponseHeaderTimeout: transportConf.ResponseHeaderTimeout,
		ExpectContinueTimeout: transportConf.ExpectContinueTimeout,
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,
	}
}
