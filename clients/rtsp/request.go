package rtsp

import (
	"fmt"
	"net/url"
	"rtsp-inspector/auth"
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
	Method      string
	URL         string
	Header      Header
	Credentials Credentials
	cseq        int
	digestAuth  auth.DigestAuth
}

func NewRequest(method, rtspURL string) (*Request, error) {
	u, err := url.Parse(rtspURL)
	if err != nil {
		return nil, err
	}

	pass, _ := u.User.Password()
	username := u.User.Username()

	clearURL, err := url.Parse(rtspURL)
	if err != nil {
		return nil, err
	}
	clearURL.User = nil

	return &Request{
		Method: method,
		URL:    clearURL.String(),
		Credentials: Credentials{
			Username: username,
			Password: pass,
		},
		Header: getDefaultHeaders(),
		digestAuth: auth.DigestAuth{
			User: username,
			Pass: pass,
			URL:  rtspURL,
		},
	}, nil
}

func getDefaultHeaders() Header {
	header := make(Header)
	header.Add("User-Agent", "RTSP-Inspector")
	return header
}

func (r *Request) AddCSeq() {
	r.cseq++
	r.Header.Add("CSeq", strconv.Itoa(r.cseq))
}

func (r *Request) SetNonce(nonce string) {
	r.digestAuth.Nonce = nonce
}

func (r *Request) SetRealm(realm string) {
	r.digestAuth.Realm = realm
}

func (r *Request) Build() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s RTSP/1.0", r.Method, r.URL))
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

	if r.digestAuth.Realm != "" {
		b.WriteString(fmt.Sprintf("Authorization: %s", r.digestAuth.GetHeader(r.Method)))
		b.WriteString("\r\n")
	}
	b.WriteString("\r\n")

	return b.String()
}
