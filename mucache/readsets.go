package mucache

import (
	"demo/u"
)

// Set 类型和相关方法

type Set map[string]struct{}

func NewSet() Set {
	return make(Set)
}

func (s Set) Add(value string) {
	s[value] = struct{}{}
}

func (s Set) Contains(value string) bool {
	_, exists := s[value]
	return exists
}

func (s Set) Size() int {
	return len(s)
}

// ReadSets 是一个map，其键是string，值是Set

type ReadSets map[string]Set

func NewReadSets() ReadSets {
	return make(ReadSets)
}

func (rs ReadSets) Add(key u.CallArgs, value interface{}) {
	strKey := key.String()
	if _, exists := rs[strKey]; !exists {
		rs[strKey] = NewSet()
	}
	switch v := value.(type) {
	case u.CallArgs:
		rs[strKey].Add(v.String())
	case string:
		rs[strKey].Add(v)

	}

}

func (rs ReadSets) Contains(key u.CallArgs, value interface{}) bool {
	strKey := key.String()
	if set, exists := rs[strKey]; exists {
		switch v := value.(type) {
		case u.CallArgs:
			return set.Contains(v.String())
		case string:
			return set.Contains(v)
		default:
			return false
		}
	}
	return false
}

func (rs ReadSets) Pop(key u.CallArgs) (Set, bool) {
	strKey := key.String()
	if set, exists := rs[strKey]; exists {
		delete(rs, strKey) // 从ReadSets中删除该key
		return set, true   // 返回被删除的Set和true，表示成功弹出
	}
	return nil, false // 如果key不存在，返回nil和false
}
