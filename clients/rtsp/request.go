package rtsp

import (
	"fmt"
	"rtsp-inspector/types"
	"strconv"
	"strings"
)

type Credentials struct {
	Username string
	Password string
}

type Header map[string]string

func (h Header) Add(key, value string) {
	h[key] = value
}

type Request struct {
	Method  string
	URL     string
	Header  Header
	TrackID int
}

func (c *Client) NewRequest(method, rtspURL string) (*Request, error) {
	return &Request{
		Method: method,
		URL:    rtspURL,
		Header: getDefaultHeaders(),
	}, nil
}

func getDefaultHeaders() Header {
	header := make(Header)
	header.Add("User-Agent", "RTSP-Inspector")
	return header
}

func (c *Client) BuildRequest(r Request) string {
	r.Header.Add("CSeq", strconv.Itoa(c.csec))

	if c.digestAuth.Nonce != "" {
		r.Header.Add("Authorization", c.digestAuth.GetHeader(r.Method, r.URL))
	}

	var b strings.Builder

	url := r.URL

	b.WriteString(fmt.Sprintf("%s %s RTSP/1.0", r.Method, url))
	b.WriteString("\r\n")

	for key, value := range r.Header {
		b.WriteString(fmt.Sprintf("%s: %s", key, value))
		b.WriteString("\r\n")
	}

	if r.Method == types.MethodSetup {
		if _, ok := r.Header["Transport"]; !ok {
			b.WriteString("Transport: RTP/AVP/TCP;interleaved=0-1")
			b.WriteString("\r\n")
		}
	}

	if r.Method == types.MethodDescribe {
		if _, ok := r.Header["Accept"]; !ok {
			b.WriteString("Accept: application/sdp")
			b.WriteString("\r\n")
		}
	}

	b.WriteString("\r\n")

	return b.String()
}
