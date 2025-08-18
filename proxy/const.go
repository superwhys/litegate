// File:		const.go
// Created by:	Hoven
// Created on:	2025-08-15
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package proxy

import "time"

type claimContextKey string

const (
	PlaceHeader = "$header"
	PlaceQuery  = "$query"

	DefaultTimeout = 30 * time.Second
)
