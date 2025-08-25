// File:		buffer_test.go
// Created by:	Hoven
// Created on:	2025-08-26
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.
package agent

import (
	"testing"
)

func TestBufferPool_Basic(t *testing.T) {
	// 创建缓冲区池
	bp := newBufferPool(1024)

	// 测试获取缓冲区
	buf1 := bp.Get()
	if len(buf1) != 0 {
		t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(buf1))
	}
	if cap(buf1) != 1024 {
		t.Errorf("期望缓冲区容量为1024，实际为%d", cap(buf1))
	}

	// 测试使用缓冲区
	data := []byte("hello world")
	buf1 = append(buf1, data...)
	if len(buf1) != len(data) {
		t.Errorf("期望缓冲区长度为%d，实际为%d", len(data), len(buf1))
	}

	// 测试放回缓冲区
	bp.Put(buf1)

	// 再次获取，应该得到重置的缓冲区
	buf2 := bp.Get()
	if len(buf2) != 0 {
		t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(buf2))
	}
	if cap(buf2) != 1024 {
		t.Errorf("期望缓冲区容量为1024，实际为%d", cap(buf2))
	}
}

func TestBufferPool_DifferentSizes(t *testing.T) {
	// 测试不同大小的缓冲区池
	testCases := []struct {
		poolSize   int
		bufferSize int
	}{
		{1, 512},
		{5, 1024},
		{10, 2048},
		{100, 4096},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			bp := newBufferPool(tc.bufferSize)

			// 获取缓冲区
			buf := bp.Get()
			if len(buf) != 0 {
				t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(buf))
			}
			if cap(buf) != tc.bufferSize {
				t.Errorf("期望缓冲区容量为%d，实际为%d", tc.bufferSize, cap(buf))
			}

			// 放回缓冲区
			bp.Put(buf)
		})
	}
}

func TestBufferPool_Reuse(t *testing.T) {
	bp := newBufferPool(1024)

	// 获取多个缓冲区
	buffers := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		buffers[i] = bp.Get()
		// 写入一些数据
		buffers[i] = append(buffers[i], []byte("test data")...)
	}

	// 放回所有缓冲区
	for _, buf := range buffers {
		bp.Put(buf)
	}

	// 再次获取，验证缓冲区被正确重置
	for i := 0; i < 10; i++ {
		buf := bp.Get()
		if len(buf) != 0 {
			t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(buf))
		}
		if cap(buf) != 1024 {
			t.Errorf("期望缓冲区容量为1024，实际为%d", cap(buf))
		}
	}
}

func TestBufferPool_WrongSize(t *testing.T) {
	bp := newBufferPool(1024)

	// 获取一个缓冲区
	buf := bp.Get()

	// 创建一个不同大小的缓冲区
	wrongBuf := make([]byte, 512)

	// 放回错误大小的缓冲区（应该被丢弃）
	bp.Put(wrongBuf)

	// 放回正确大小的缓冲区
	bp.Put(buf)

	// 再次获取，应该正常工作
	newBuf := bp.Get()
	if len(newBuf) != 0 {
		t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(newBuf))
	}
	if cap(newBuf) != 1024 {
		t.Errorf("期望缓冲区容量为1024，实际为%d", cap(newBuf))
	}
}

func TestBufferPool_Concurrent(t *testing.T) {
	bp := newBufferPool(1024)

	// 并发测试
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// 获取缓冲区
			buf := bp.Get()

			// 写入数据
			buf = append(buf, []byte("concurrent test")...)

			// 放回缓冲区
			bp.Put(buf)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证池仍然正常工作
	buf := bp.Get()
	if len(buf) != 0 {
		t.Errorf("期望获取的缓冲区长度为0，实际为%d", len(buf))
	}
	if cap(buf) != 1024 {
		t.Errorf("期望缓冲区容量为1024，实际为%d", cap(buf))
	}
}

func BenchmarkBufferPool_GetPut(b *testing.B) {
	bp := newBufferPool(1024)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bp.Get()
			bp.Put(buf)
		}
	})
}

func BenchmarkBufferPool_WithData(b *testing.B) {
	bp := newBufferPool(1024)
	testData := []byte("test data for benchmark")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bp.Get()
			buf = append(buf, testData...)
			bp.Put(buf)
		}
	})
}
