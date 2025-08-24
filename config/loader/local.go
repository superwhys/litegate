package loader

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/litegate/config"
)

type localConfigLoader struct {
	configDir string

	routeConfigs map[string]*config.RouteConfig
	watcher      *fsnotify.Watcher
	stopChan     chan struct{}
	mu           sync.RWMutex
}

func NewLocalConfigLoader(configDir string) *localConfigLoader {
	ll := &localConfigLoader{
		configDir:    configDir,
		routeConfigs: make(map[string]*config.RouteConfig),
		stopChan:     make(chan struct{}),
	}

	ll.loadAllConfigs()

	return ll
}

func (ll *localConfigLoader) loadAllConfigs() {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if err := os.MkdirAll(ll.configDir, 0755); err != nil {
		logging.Errorf("create config dir error: %v", err)
		return
	}

	err := filepath.WalkDir(ll.configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logging.Errorf("walk config dir error: %v", err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !ll.isConfigFile(path) {
			return nil
		}

		if err := ll.loadConfigFile(path); err != nil {
			logging.Errorf("load config file error: %s, %v", path, err)
			return nil
		}

		return nil
	})

	if err != nil {
		logging.Errorf("load config file error: %v", err)
	}

	logging.Infof("load %d config files", len(ll.routeConfigs))
}

func (ll *localConfigLoader) loadConfigFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var routeConfig config.RouteConfig
	if err := json.Unmarshal(data, &routeConfig); err != nil {
		return err
	}

	serviceName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	ll.routeConfigs[serviceName] = &routeConfig

	logging.Infof("load config file: %s -> %s", filePath, serviceName)
	return nil
}

func (ll *localConfigLoader) Get(service string) (*config.RouteConfig, error) {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	return ll.routeConfigs[service], nil
}

func (ll *localConfigLoader) GetAll() ([]*config.RouteConfig, error) {
	ll.mu.RLock()
	defer ll.mu.RUnlock()

	routeConfigs := make([]*config.RouteConfig, 0, len(ll.routeConfigs))
	for _, routeConfig := range ll.routeConfigs {
		routeConfigs = append(routeConfigs, routeConfig)
	}
	return routeConfigs, nil
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

	logging.Infof("start watch config dir: %s", ll.configDir)
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
		logging.Infof("stop watch config dir: %s", ll.configDir)
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
			logging.Errorf("watch config dir error: %v", err)
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
		logging.Infof("config file changed: %s", event.Name)
		ll.onConfigChanged(event.Name)
	case event.Has(fsnotify.Create):
		logging.Infof("config file created: %s", event.Name)

		// 如果是新创建的目录，添加到监听列表
		if ll.isDirectory(event.Name) {
			ll.watcher.Add(event.Name)
		}
		ll.onConfigChanged(event.Name)
	case event.Has(fsnotify.Remove):
		logging.Infof("config file removed: %s", event.Name)
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
	if err := ll.loadConfigFile(filename); err != nil {
		logging.Errorf("reload config file error: %s, %v", filename, err)
	}
}

func (ll *localConfigLoader) onConfigRemoved(filename string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	serviceName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	delete(ll.routeConfigs, serviceName)
	logging.Infof("remove config file: %s -> %s", filename, serviceName)
}
