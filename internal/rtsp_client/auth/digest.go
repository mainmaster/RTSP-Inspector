package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

const digestHeader = `Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`

type DigestAuth struct {
	Username, Password, Realm, Nonce string
}

func (a *DigestAuth) Calc(method, url string) string {
	ha1 := md5Sum(fmt.Sprintf("%s:%s:%s", a.Username, a.Realm, a.Password))
	ha2 := md5Sum(fmt.Sprintf("%s:%s", method, url))
	return md5Sum(fmt.Sprintf("%s:%s:%s", ha1, a.Nonce, ha2))
}

func (a *DigestAuth) GetHeader(method, url string) string {
	return fmt.Sprintf(digestHeader, a.Username, a.Realm, a.Nonce, url, a.Calc(method, url))
}

func md5Sum(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}
