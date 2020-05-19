package sessions

import (
	"github.com/go-redis/redis/v7"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	cookieName string
	mutex      sync.Mutex
	expire     int64
	Provider   Provider
}

var (
	GSessions     *Manager
	StoreInstance Provider
)

func NewManager(cookieName string, expire int64) *Manager {
	return &Manager{cookieName: cookieName, expire: expire}
}

func InitRedisSessions(cookieName string, expire int64, conn *redis.Client) {
	GSessions = &Manager{
		cookieName: cookieName,
		expire:     expire,
		Provider:   NewRedisStore(conn, time.Duration(expire)*time.Second),
	}
	StoreInstance = GSessions.Provider
}

func SessionStart(w http.ResponseWriter, r *http.Request) (*Session, error) {
	return GSessions.SessionStart(w, r)
}

func (ss *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (*Session, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	cookie, err := r.Cookie(ss.cookieName)
	var session *Session
	if err != nil || cookie.Value == "" {
		sid := ss.generateSessionID()
		session, _ = ss.Provider.SessionInit(sid)
		cookie := http.Cookie{
			Name:     ss.cookieName,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
	} else {
		session, err = ss.Provider.SessionRead(cookie.Value)
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (ss *Manager) generateSessionID() string {
	return GenerateRandString(24)
}
