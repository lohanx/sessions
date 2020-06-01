package sessions

import (
        "fmt"
        "sync"
        "time"
)

type Session struct {
        sid          string
        values       map[interface{}]interface{}
        expire       time.Duration
        mutex        sync.RWMutex
        destroyState bool
        changed      bool
}

func NewSession(sid string, values map[interface{}]interface{}, expire time.Duration) *Session {
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
func (s *Session) Get(key interface{}) interface{} {
        s.mutex.RLock()
        defer s.mutex.RUnlock()
        if value, ok := s.values[key]; ok {
                return value
        }
        return nil
}

func (s *Session) Set(key, value interface{}) {
        s.mutex.Lock()
        s.destroyState = false
        s.values[key] = value
        s.changed = true
        s.mutex.Unlock()
}

func (s *Session) Delete(key interface{}) {
        s.mutex.Lock()
        if _, ok := s.values[key]; ok {
                delete(s.values, key)
        }
        s.changed = true
        s.mutex.Unlock()
}

func (s *Session) SetFlush(key interface{}, value interface{}) {
        s.mutex.Lock()
        s.destroyState = false
        s.values[s.flushKey(key)] = value
        s.changed = true
        s.mutex.Unlock()
}

func (s *Session) GetFlush(key interface{}) interface{} {
        s.mutex.Lock()
        defer s.mutex.Unlock()
        key = s.flushKey(key)
        if value, ok := s.values[key]; ok {
                delete(s.values, key)
                s.changed = true
                return value
        }
        return nil
}

func (s *Session) Has(key interface{}) bool {
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

func (s *Session) GetValues() map[interface{}]interface{} {
        return s.values
}

func (s *Session) GetDestroyState() bool {
        return s.destroyState
}

func (s *Session) GetChanged() bool {
        return s.changed
}

func (s *Session) GetExpire() time.Duration {
        return s.expire
}

func (s *Session) SetExpire(expire int64) {
        s.expire = time.Duration(expire) * time.Second
}

func (s *Session) flushKey(key interface{}) string {
        return fmt.Sprintf("__flush.%+v", key)
}