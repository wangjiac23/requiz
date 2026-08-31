// 由 server.go 拆分生成（V1.6.0 模块化，功能零变化）
package web

import (
	"requiz/src/model"
	"requiz/src/config"
	"requiz/src/quiz"
	"fmt"
	"path/filepath"
	"sync"
)

type Store struct {
	mu     sync.Mutex
	main   *model.Bank
	links  []*model.Bank
	global model.GlobalConfig // V1.3.0 全局配置（用户级）
}


func newStore(main *model.Bank) *Store {
	s := &Store{main: main}
	// V1.3.0：读全局配置（首次自动创建默认模板），主题库自动注册（Obsidian 式：打开即记录）
	gc, err := config.ReadGlobalConfig()
	if err == nil {
		s.global = gc
		// 主题库自动加入全局题库列表
		if !containsStr(gc.Links, main.Dir) {
			gc.Links = append(gc.Links, main.Dir)
			_ = config.WriteGlobalConfig(gc)
		}
		s.global = gc
		migrateProjectLinks(main, &s.global)
	} else {
		s.global = config.DefaultGlobalConfig()
	}
	// 加载全局配置中的全部题库（对等：都加载，主题库跳过重复）
	for _, bankDir := range s.global.Links {
		if bankDir == main.Dir {
			continue
		}
		if b, err := quiz.ConnectBank(bankDir); err == nil {
			s.links = append(s.links, b)
		}
	}
	return s
}

// migrateProjectLinks 旧版：项目配置中的 links 迁移到全局配置

func migrateProjectLinks(main *model.Bank, gc *model.GlobalConfig) {
	if len(main.Links) == 0 {
		return
	}
	changed := false
	for _, l := range main.Links {
		if !containsStr(gc.Links, l) {
			gc.Links = append(gc.Links, l)
			changed = true
		}
	}
	if changed {
		_ = config.WriteGlobalConfig(*gc)
	}
	// 重写项目配置（去除旧版 links 节）
	pc, err := config.ReadProjectConfig(main.Dir)
	if err == nil {
		_ = config.WriteProjectConfig(main.Dir, pc)
	}
}


func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}


func (s *Store) banks() []*model.Bank {
	banks := []*model.Bank{s.main}
	return append(banks, s.links...)
}

// bankByDir 按目录找题库；dir 为空返回主题库；也支持按题库名匹配

func (s *Store) bankByDir(dir string) (*model.Bank, error) {
	if dir == "" {
		return s.main, nil
	}
	for _, b := range s.banks() {
		if b.Dir == dir {
			return b, nil
		}
	}
	for _, b := range s.banks() {
		if b.Name == dir {
			return b, nil
		}
	}
	return nil, fmt.Errorf("未链接的题库: %s", dir)
}

// addLink 连接新题库并持久化到全局配置（V1.3.0：links 存用户级全局配置）

func (s *Store) addLink(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if containsStr(s.global.Links, abs) {
		return nil // 已链接
	}
	b, err := quiz.ConnectBank(abs)
	if err != nil {
		return err
	}
	gc := s.global
	if !containsStr(gc.Links, abs) {
		gc.Links = append(gc.Links, abs)
	}
	if err := config.WriteGlobalConfig(gc); err != nil {
		return err
	}
	s.mu.Lock()
	s.global.Links = append(s.global.Links, abs)
	s.links = append(s.links, b)
	s.mu.Unlock()
	return nil
}

// ---------- Web 服务 ----------

