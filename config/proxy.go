// File:		proxy.go
// Created by:	Hoven
// Created on:	2025-08-16
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package config

import (
	"fmt"
	"math/rand"
	"time"
)

type ProxyConfigLoader interface {
	Get(service string) (*RouteConfig, error)
	GetAll() ([]*RouteConfig, error)
	Watch()
}

// Config 主配置结构体
type RouteConfig struct {
	// 代理地址列表（必填）
	Proxy []string `json:"proxy" validate:"required"`
	// 超时时间（选填，默认30秒）
	Timeout int `json:"timeout"`
	// 身份验证配置（选填）
	Auth *Auth `json:"auth,omitempty"`
	// 路由配置（必填）
	Routes []Route `json:"routes" validate:"required,min=1"`
	// 版本号（用于乐观锁）
	Version int32 `json:"version"`
}

// Auth 身份验证配置
type Auth struct {
	// Token类型，固定为jwt
	Type string `json:"type" validate:"required,eq=jwt"`
	// Token在请求中的位置
	// $header.token
	// $query.token
	Source string `json:"source" validate:"required"`
	// JWT密钥
	Secret string `json:"secret" validate:"required"`
	// JWT解码后数据存储位置映射
	// {
	// 	"$query.user_id": "user_id",
	// 	"$header.X-User": "userName",
	// }
	// 该示例表示
	// 1. 将JWT解码后的 `user_id` 数据存储到请求 query中, key为 user_id
	// 2. 将JWT解码后的 `userName` 数据存储到请求 header中, key为 X-User
	Claims map[string]string `json:"claims" validate:"required"`
}

// Route 路由配置
type Route struct {
	// URL匹配 (正则表达式)
	Match string `json:"match" validate:"required"`
	// 代理配置（单个地址或地址数组）
	Proxy ProxyConfig `json:"proxy"`
	// 超时时间（选填）
	Timeout int `json:"timeout"`
	// 是否禁用身份验证
	DisableAuth bool `json:"disable_auth"`
	// 身份验证配置覆盖
	Auth *Auth `json:"auth,omitempty"`
}

// ProxyConfig 代理配置，支持单个地址或地址数组
type ProxyConfig struct {
	// 单个地址
	Address string
	// 地址数组（用于负载均衡）
	Addresses []string
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (p ProxyConfig) pickAddress() (string, error) {
	if p.Address != "" {
		return p.Address, nil
	}
	if len(p.Addresses) > 0 {
		if len(p.Addresses) == 1 {
			return p.Addresses[0], nil
		}
		return p.Addresses[rand.Intn(len(p.Addresses))], nil
	}
	return "", fmt.Errorf("empty downstream addresses")
}

type localConfigLoader struct {
	path string
}

func NewLocalConfigLoader(path string) *localConfigLoader {
	return &localConfigLoader{path}
}

func (ll *localConfigLoader) Get(service string) (*RouteConfig, error) {
	return nil, nil
}

func (ll *localConfigLoader) GetAll(service string) ([]*RouteConfig, error) {
	return nil, nil
}

func (ll *localConfigLoader) Watch() {

}
