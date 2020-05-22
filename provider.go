package sessions

type Provider interface {
        SessionInit(sid string) (*Session, error)
        SessionRead(sid string) (*Session, error)
        SessionDestroy(sid string) error
        SessionWrite(sess *Session) error
        SessionGC()
}
