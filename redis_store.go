package sessions

import (
        "context"
        "github.com/go-redis/redis/v8"
        "time"
)

type RedisStore struct {
        conn    *redis.Client
        expire  time.Duration
        timeout time.Duration
}

func NewRedisStore(conn *redis.Client, expire, timeout time.Duration) *RedisStore {
        return &RedisStore{conn: conn, expire: expire, timeout: timeout}
}

//Initializes a new session
func (store *RedisStore) SessionInit(sid string) (*Session, error) {
        return NewSession(sid, make(map[string]interface{}), store.expire), nil
}

//Read an existing session
func (store *RedisStore) SessionRead(sid string) (*Session, error) {
        ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
        defer cancel()
        data, err := store.conn.Get(ctx, sid).Bytes()
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
        return sess, store.conn.Expire(ctx, sid, store.expire).Err()
}

//Destroy the session
func (store *RedisStore) SessionDestroy(sid string) error {
        ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
        defer cancel()
        return store.conn.Del(ctx, sid).Err()
}

//automatic
func (store *RedisStore) SessionGC() {
}

//Writes session data after the request is completed
func (store *RedisStore) SessionWrite(sess *Session) error {
        ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
        defer cancel()
        if sess.destroyState {
                return store.conn.Expire(ctx, sess.SessionID(), -time.Second*2).Err()
        }
        if !sess.changed {
                return store.conn.Expire(ctx, sess.SessionID(), sess.expire).Err()
        }
        data, err := marshal(&sess.values)
        if err != nil {
                return err
        }
        return store.conn.Set(ctx, sess.sid, data, sess.expire).Err()
}
