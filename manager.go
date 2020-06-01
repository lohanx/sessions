package sessions

import (
        "bytes"
        "encoding/base64"
        "encoding/gob"
        "fmt"
        "github.com/go-redis/redis/v8"
        uuid "github.com/satori/go.uuid"
        "github.com/vmihailenco/msgpack/v4"
        "net/http"
        "sync"
        "time"
        "unsafe"
)

type Manager struct {
        cookieName string
        mutex      sync.Mutex
        expire     int64
        Provider   Provider
        Serializer string // msgpack or gob
}

var GlobalManager *Manager

func NewManager(cookieName string, expire int64) *Manager {
        return &Manager{cookieName: cookieName, expire: expire, Serializer: "msgpack"}
}

func UseSerializer(t string) {
        GlobalManager.Serializer = t
}

func NewManagerWithRedis(cookieName string, expire, timeout int64, conn *redis.Client) {
        GlobalManager = NewManager(cookieName, expire)
        GlobalManager.Provider = NewRedisStore(conn, time.Duration(expire)*time.Second, time.Duration(timeout)*time.Second)
}

func SessionStart(w http.ResponseWriter, r *http.Request) (*Session, error) {
        return GlobalManager.SessionStart(w, r)
}

func SessionSave(session *Session) error {
        return GlobalManager.Provider.SessionWrite(session)
}

func (ss *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (*Session, error) {
        ss.mutex.Lock()
        defer ss.mutex.Unlock()
        cookie, err := r.Cookie(ss.cookieName)
        var session *Session
        if err != nil || cookie.Value == "" {
                sid := ss.generate()
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

func marshal(v interface{}) ([]byte, error) {
        switch GlobalManager.Serializer {
        case "gob":
                buff := bytes.Buffer{}
                err := gob.NewEncoder(&buff).Encode(v)
                return buff.Bytes(), err
        case "msgpack":
                return msgpack.Marshal(v)
        }
        return nil, fmt.Errorf("unsupported serializer")
}

func unmarshal(data []byte, v interface{}) error {
        switch GlobalManager.Serializer {
        case "gob":
                buff := bytes.NewReader(data)
                return gob.NewDecoder(buff).Decode(v)
        case "msgpack":
                return msgpack.Unmarshal(data, v)
        }
        return fmt.Errorf("unsupported serializer")
}

func (ss *Manager) generate() string {
        data := uuid.NewV4().Bytes()
        std := base64.StdEncoding
        buff := make([]byte, std.EncodedLen(len(data)))
        std.Encode(buff, data)
        for i, b := range buff {
                if b == '/' {
                        buff[i] = '_'
                }
        }
        buff = buff[:22]
        return *(*string)(unsafe.Pointer(&buff))
}
