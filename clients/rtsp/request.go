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

func NewRequest(method, rtspURL string) (*Request, error) {
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

func (r *Request) SetCSeq(cseq int) {
	r.Header.Add("CSeq", strconv.Itoa(cseq))
}

func (r *Request) Build() string {
	var b strings.Builder

	url := r.URL
	if r.Method == types.MethodSetup {
		url = fmt.Sprintf("%s/trackID=%d", url, r.TrackID)
	}

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
