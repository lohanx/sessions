package sessions

import (
	"github.com/go-redis/redis/v7"
	"time"
)

type RedisStore struct {
	conn   *redis.Client
	expire time.Duration
}

func NewRedisStore(conn *redis.Client, expire time.Duration) *RedisStore {
	return &RedisStore{conn: conn, expire: expire}
}

//Initializes a new session
func (store *RedisStore) SessionInit(sid string) (*Session, error) {
	return NewSession(sid, make(map[string]interface{}), store.expire), nil
}

//Read an existing session
func (store *RedisStore) SessionRead(sid string) (*Session, error) {
	data, err := store.conn.Get(sid).Bytes()
	if err != nil {
		if err == redis.Nil {
			return store.SessionInit(sid)
		}
		return nil, err
	}
	sess := NewSession(sid, make(map[string]interface{}), store.expire)
	if err = unmarshal(data, &sess.values); err != nil {
		return nil, err
	}
	return sess, store.conn.Expire(sid, store.expire).Err()
}

//Destroy the session
func (store *RedisStore) SessionDestroy(sid string) error {
	return store.conn.Del(sid).Err()
}

//automatic
func (store *RedisStore) SessionGC() {
}

//Writes session data after the request is completed
func (store *RedisStore) SessionWrite(sess *Session) error {
	if sess.destroyState {
		return store.conn.Expire(sess.SessionID(), -time.Second*2).Err()
	}
	if !sess.changed {
		return nil
	}
	data, err := marshal(&sess.values)
	if err != nil {
		return err
	}
	return store.conn.Set(sess.sid, data, sess.expire).Err()
}
