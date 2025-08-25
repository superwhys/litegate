// File:		buffer.go
// Created by:	Hoven
// Created on:	2025-08-26
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.
package agent

import (
	"sync"
)

type bufferPool struct {
	pool *sync.Pool
	size int
}

func newBufferPool(bufferSize int) *bufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() any {
				buf := make([]byte, bufferSize)
				return &buf
			},
		},
		size: bufferSize,
	}
}

func (bp *bufferPool) Get() []byte {
	bufPtr := bp.pool.Get().(*[]byte)
	return (*bufPtr)[:0]
}

func (bp *bufferPool) Put(buf []byte) {
	if cap(buf) == bp.size {
		buf = buf[:0]
		bp.pool.Put(&buf)
	}
}
