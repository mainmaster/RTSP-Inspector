package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

const digestHeader = `Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`

type DigestAuth struct {
	User, Pass, Realm, Nonce, URL string
}

func (a *DigestAuth) Calc(method string) string {
	ha1 := md5Sum(fmt.Sprintf("%s:%s:%s", a.User, a.Realm, a.Pass))
	ha2 := md5Sum(fmt.Sprintf("%s:%s", method, a.URL))
	return md5Sum(fmt.Sprintf("%s:%s:%s", ha1, a.Nonce, ha2))
}

func (a *DigestAuth) GetHeader(method string) string {
	return fmt.Sprintf(digestHeader, a.User, a.Realm, a.Nonce, a.URL, a.Calc(method))
}

func md5Sum(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}
