package types

import (
	"net/textproto"
)

type Headers struct {
	StatusCode int
	StatusLine string
	Header     textproto.MIMEHeader
}

type Response struct {
	Headers
	Body []byte
}
