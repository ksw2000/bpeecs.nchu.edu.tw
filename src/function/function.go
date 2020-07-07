package function

import(
    "math/rand"
    "time"
    "net/http"
    "strings"
)

func RandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~@!"
    var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
    b := make([]byte, length)
        for i := range b {
            b[i] = charset[seededRand.Intn(len(charset))]
        }
    return string(b)
}

func GET(key string, r *http.Request) string{
    return strings.Join(r.Form[key], "")
}
