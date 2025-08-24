package loader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalConfigLoader_Watch(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建配置加载器
	loader := NewLocalConfigLoader(tempDir)

	// 启动文件监听
	if err := loader.Watch(); err != nil {
		t.Fatalf("启动文件监听失败: %v", err)
	}

	// 等待一下确保监听器启动
	time.Sleep(100 * time.Millisecond)

	// 创建测试配置文件
	testConfigFile := filepath.Join(tempDir, "test.json")
	testContent := `{"proxy": ["http://localhost:8080"], "routes": [{"match": "/test"}]}`

	if err := os.WriteFile(testConfigFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 等待文件创建事件处理
	time.Sleep(100 * time.Millisecond)

	// 修改配置文件
	modifiedContent := `{"proxy": ["http://localhost:8081"], "routes": [{"match": "/test"}]}`
	if err := os.WriteFile(testConfigFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("修改测试配置文件失败: %v", err)
	}

	// 等待文件修改事件处理
	time.Sleep(100 * time.Millisecond)

	// 删除配置文件
	if err := os.Remove(testConfigFile); err != nil {
		t.Fatalf("删除测试配置文件失败: %v", err)
	}

	// 等待文件删除事件处理
	time.Sleep(100 * time.Millisecond)

	// 停止监听
	loader.StopWatch()

	t.Log("文件监听测试完成")
}

func TestLocalConfigLoader_SubdirWatch(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建子目录
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}

	// 创建配置加载器
	loader := NewLocalConfigLoader(tempDir)

	// 启动文件监听
	if err := loader.Watch(); err != nil {
		t.Fatalf("启动文件监听失败: %v", err)
	}

	// 等待一下确保监听器启动
	time.Sleep(100 * time.Millisecond)

	// 在子目录中创建配置文件
	testConfigFile := filepath.Join(subDir, "test.yaml")
	testContent := `proxy: ["http://localhost:8080"]`

	if err := os.WriteFile(testConfigFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("在子目录创建测试配置文件失败: %v", err)
	}

	// 等待文件创建事件处理
	time.Sleep(100 * time.Millisecond)

	// 停止监听
	loader.StopWatch()

	t.Log("子目录文件监听测试完成")
}
