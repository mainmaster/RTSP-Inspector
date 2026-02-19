package rtsp

import (
	"fmt"
	"net/url"
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
	URL     url.URL
	Header  Header
	TrackID int
	client  *Client
}

func (c *Client) NewRequest(method, rtspURL string) (*Request, error) {
	parsedURL, err := url.Parse(rtspURL)
	if err != nil {
		return nil, err
	}
	parsedURL.User = nil

	return &Request{
		Method: method,
		URL:    *parsedURL,
		Header: getDefaultHeaders(),
		client: c,
	}, nil
}

func getDefaultHeaders() Header {
	header := make(Header)
	header.Add("User-Agent", "RTSP-Inspector")
	return header
}

func (r *Request) SetSessionID(sessionID string) {
	s := strings.Split(sessionID, ";")[0]
	r.Header.Add("Session", s)
}

func (r *Request) BuildRequest() string {
	r.Header.Add("CSeq", strconv.Itoa(r.client.csec))

	if r.client.digestAuth.Nonce != "" {
		r.Header.Add("Authorization", r.client.digestAuth.GetHeader(r.Method, r.URL.String()))
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s RTSP/1.0", r.Method, r.URL.String()))
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

	fmt.Println()
	return b.String()
}
