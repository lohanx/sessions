###go session

- gin
```go
    sessions.InitRedisSessions("session_name",1800,redis.NewClint(&redis.Options{
        Addr: "127.0.0.1:6379",
        DB: 0,
    }))
    app := gin.New()
    app.Use(func(ctx *gin.Context) {
        session,_ := sessions.SessionStart(ctx.Write,ctx.Request)
        ctx.Set("session",session)
        defer func(session *sessions.Session) {
            sessions.StoreInstance.SessionWrite(session)
        }(session)
        ctx.Next()
    })
    app.GET("/",Home)
    ...
```
handler
```go
    func Home(ctx *gin.Context) {
        sess := ctx.MustGet("session").(*sessions.Session)
        sess.Set("key","value")
        sess.Get("key")
        ...
    }
```
