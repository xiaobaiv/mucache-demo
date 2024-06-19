package mucache

import (
	"demo/u"
	"sync"
)

type Saved struct {
	mu   sync.RWMutex // 保证并发访问的安全性
	data map[string]map[u.Service]string
}

// NewSaved 创建一个新的 Saved 实例
func NewSaved() *Saved {
	return &Saved{
		data: make(map[string]map[u.Service]string),
	}
}

// Add 方法添加一个新的键值、服务和调用参数
func (s *Saved) Add(key string, service u.Service, ca u.CallArgs) {
	s.mu.Lock() // 写操作上锁
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		s.data[key] = make(map[u.Service]string) // 如果键不存在，初始化内层map
	}
	s.data[key][service] = ca.String() // 存储 Service 和 CallArgs
}

// PopByKey 方法按照 key 值取出并删除对应的 service 和 CallArgs
func (s *Saved) PopByKey(key string) (map[u.Service]string, bool) {
	s.mu.Lock() // 写操作上锁
	defer s.mu.Unlock()

	// 检查 key 是否存在
	serviceCallArgsMap, ok := s.data[key]
	if !ok {
		return nil, false // 没有找到对应的 key
	}

	// 删除 key 并返回对应的 service 和 CallArgs
	delete(s.data, key)
	return serviceCallArgsMap, true
}
