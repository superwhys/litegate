// File:		server.go
// Created by:	Hoven
// Created on:	2025-08-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package config

import "time"

type TransportConfig struct {
	MaxIdleConns          int           `json:"max_idle_conns"`          // 最大空闲连接数
	MaxIdleConnsPerHost   int           `json:"max_idle_conns_per_host"` // 每个主机的最大空闲连接数
	MaxConnsPerHost       int           `json:"max_conns_per_host"`      // 每个主机的最大连接数
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`       // 空闲连接超时时间
	FlushInterval         time.Duration `json:"flush_interval"`          // 刷新间隔
	BufferSize            int           `json:"buffer_size"`             // 单个缓冲区大小
	TcpDialTimeout        time.Duration `json:"tcp_dial_timeout"`        // TCP连接超时时间
	TcpHandshakeTimeout   time.Duration `json:"tcp_handshake_timeout"`   // TCP握手超时时间
	KeepAlive             time.Duration `json:"keep_alive"`              // 保持连接超时时间
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"` // 响应头超时时间
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"` // 期望继续超时时间
}

type GatewayConfig struct {
	// Services list which allowed to be accessed
	Services  []string         `json:"services"`
	Timeout   time.Duration    `json:"timeout"`
	Transport *TransportConfig `json:"transport"`
}

func (c *GatewayConfig) SetDefault() {
	if c.Transport == nil {
		c.Transport = &TransportConfig{
			MaxIdleConns:          500,
			MaxIdleConnsPerHost:   50,
			MaxConnsPerHost:       500,
			IdleConnTimeout:       60 * time.Second,
			FlushInterval:         10 * time.Millisecond,
			BufferSize:            128 * 1024,
			TcpDialTimeout:        5 * time.Second,
			TcpHandshakeTimeout:   5 * time.Second,
			KeepAlive:             30 * time.Second,
			ResponseHeaderTimeout: 4 * time.Second,
			ExpectContinueTimeout: 2 * time.Second,
		}
	}

	if c.Timeout == 0 {
		c.Timeout = 15 * time.Second
	}
}
