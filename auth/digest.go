package auth

import (
	"crypto/md5"
	"fmt"
)

type DigestAuth struct {
	User, Pass, Realm, Nonce, URL string
}

func (a *DigestAuth) Calc(method string) string {
	ha1 := fmt.Sprintf("%x", md5.Sum([]byte(a.User+":"+a.Realm+":"+a.Pass)))
	ha2 := fmt.Sprintf("%x", md5.Sum([]byte(method+":"+a.URL)))
	return fmt.Sprintf("%x", md5.Sum([]byte(ha1+":"+a.Nonce+":"+ha2)))
}

func (a *DigestAuth) GetHeader(method string) string {
	return fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		a.User, a.Realm, a.Nonce, a.URL, a.Calc(method))
}
