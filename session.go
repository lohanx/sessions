package sessions

import (
	"fmt"
	"sync"
	"time"
)

type Session struct {
	sid          string
	values       map[string]interface{}
	expire       time.Duration
	mutex        sync.RWMutex
	destroyState bool
	changed      bool
}

func NewSession(sid string, values map[string]interface{}, expire time.Duration) *Session {
	return &Session{
		sid:          sid,
		values:       values,
		expire:       expire,
		destroyState: false,
		changed:      false,
	}
}

func (s *Session) SessionID() string {
	return s.sid
}

//get value
func (s *Session) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if value, ok := s.values[key]; ok {
		return value
	}
	return nil
}

func (s *Session) Set(key string, value interface{}) {
	s.mutex.Lock()
	s.destroyState = false
	s.values[key] = value
	s.changed = true
	s.mutex.Unlock()
}

func (s *Session) Delete(key string) {
	s.mutex.Lock()
	if _, ok := s.values[key]; ok {
		delete(s.values, key)
	}
	s.changed = true
	s.mutex.Unlock()
}

func (s *Session) SetFlush(key string, value interface{}) {
	s.mutex.Lock()
	s.destroyState = false
	s.values[s.flushKey(key)] = value
	s.changed = true
	s.mutex.Unlock()
}

func (s *Session) GetFlush(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	key = s.flushKey(key)
	if value, ok := s.values[key]; ok {
		delete(s.values, key)
		s.changed = true
		return value
	}
	return nil
}

func (s *Session) Has(key string) bool {
	s.mutex.RLock()
	_, ok := s.values[key]
	s.mutex.RUnlock()
	return ok
}

func (s *Session) Count() int {
	return len(s.values)
}

func (s *Session) Destroy() {
	s.mutex.Lock()
	for k := range s.values {
		delete(s.values, k)
	}
	s.destroyState = true
	s.mutex.Unlock()
}

func (s *Session) DestroyState() bool {
	return s.destroyState
}

func (s *Session) flushKey(key string) string {
	return fmt.Sprintf("__flush.%s", key)
}
