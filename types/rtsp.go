package types

import "net/textproto"

type Response struct {
	StatusCode int
	StatusLine string
	Header     textproto.MIMEHeader
	Body       []byte
}
