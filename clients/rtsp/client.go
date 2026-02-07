package rtsp

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"rtsp-inspector/auth"
)

type Client struct {
	digestAuth auth.DigestAuth
	rtspURL    *url.URL
	conn       net.Conn
	reader     *bufio.Reader
	tp         *textproto.Reader
	cseq       int
}

func NewClient(rtspURL string) (*Client, error) {
	u, err := url.Parse(rtspURL)
	if err != nil {
		return nil, err
	}

	connAddr := fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
	conn, err := net.Dial("tcp", connAddr)
	if err != nil {
		return nil, err
	}

	pass, _ := u.User.Password()
	username := u.User.Username()
	u.User = nil

	r := bufio.NewReader(conn)
	return &Client{
		digestAuth: auth.DigestAuth{User: username, Pass: pass, URL: u.String()},
		rtspURL:    u,
		conn:       conn,
		reader:     r,
		tp:         textproto.NewReader(r),
	}, nil
}
