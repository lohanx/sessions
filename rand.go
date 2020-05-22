package sessions

import (
        "math/rand"
        "strings"
        "time"
)

const (
        Letters    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
        LettersLen = int32(len(Letters))
)

func init() {
        rand.Seed(time.Now().UnixNano())
}

func GenerateRandString(length int) string {
        buff := strings.Builder{}
        buff.Grow(length)
        for i := 0; i < length; i++ {
                n := rand.Int31n(LettersLen)
                buff.WriteByte(Letters[n])
        }
        return buff.String()
}
