package loader

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/litegate/config"
)

type localConfigLoader struct {
	configDir string
	watcher   *fsnotify.Watcher
	mu        sync.RWMutex
	stopChan  chan struct{}
}

func NewLocalConfigLoader(configDir string) *localConfigLoader {
	return &localConfigLoader{
		configDir: configDir,
		stopChan:  make(chan struct{}),
	}
}

func (ll *localConfigLoader) Get(service string) (*config.RouteConfig, error) {
	return nil, nil
}

func (ll *localConfigLoader) GetAll() ([]*config.RouteConfig, error) {
	return nil, nil
}

func (ll *localConfigLoader) Watch() error {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	ll.watcher = watcher

	if err := os.MkdirAll(ll.configDir, 0755); err != nil {
		watcher.Close()
		return err
	}

	if err := watcher.Add(ll.configDir); err != nil {
		watcher.Close()
		return err
	}

	if err := ll.addSubdirsToWatch(); err != nil {
		watcher.Close()
		return err
	}

	go ll.watchLoop()

	logging.Infof("开始监听配置目录: %s", ll.configDir)
	return nil
}

// StopWatch 停止文件监听
func (ll *localConfigLoader) StopWatch() {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if ll.watcher != nil {
		close(ll.stopChan)
		ll.watcher.Close()
		ll.watcher = nil
		logging.Infof("停止监听配置目录: %s", ll.configDir)
	}
}

// watchLoop 文件监听主循环
func (ll *localConfigLoader) watchLoop() {
	for {
		select {
		case event, ok := <-ll.watcher.Events:
			if !ok {
				return
			}
			ll.handleFileEvent(event)
		case err, ok := <-ll.watcher.Errors:
			if !ok {
				return
			}
			logging.Errorf("文件监听错误: %v", err)
		case <-ll.stopChan:
			return
		}
	}
}

func (ll *localConfigLoader) handleFileEvent(event fsnotify.Event) {
	if !ll.isConfigFile(event.Name) {
		return
	}

	switch {
	case event.Has(fsnotify.Write):
		logging.Infof("配置文件已修改: %s", event.Name)
		ll.onConfigChanged(event.Name)
	case event.Has(fsnotify.Create):
		logging.Infof("配置文件已创建: %s", event.Name)

		// 如果是新创建的目录，添加到监听列表
		if ll.isDirectory(event.Name) {
			ll.watcher.Add(event.Name)
		}
		ll.onConfigChanged(event.Name)
	case event.Has(fsnotify.Remove):
		logging.Infof("配置文件已删除: %s", event.Name)
		ll.onConfigRemoved(event.Name)
	}
}

func (ll *localConfigLoader) addSubdirsToWatch() error {
	return filepath.WalkDir(ll.configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != ll.configDir {
			return ll.watcher.Add(path)
		}
		return nil
	})
}

func (ll *localConfigLoader) isConfigFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".json"
}

func (ll *localConfigLoader) isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (ll *localConfigLoader) onConfigChanged(filename string) {
	// 这里可以添加配置文件重新加载的逻辑
	// 例如：重新解析配置文件、更新内存中的配置等
	logging.Infof("处理配置文件变化: %s", filename)
}

func (ll *localConfigLoader) onConfigRemoved(filename string) {
	// 这里可以添加配置文件删除后的处理逻辑
	logging.Infof("处理配置文件删除: %s", filename)
}
