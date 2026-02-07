package types

import (
	"net/textproto"

	"github.com/pixelbender/go-sdp/sdp"
)

type Headers struct {
	StatusCode int
	StatusLine string
	Header     textproto.MIMEHeader
}

type Response struct {
	Headers
	Body       []byte
	SDPSession *sdp.Session
}
