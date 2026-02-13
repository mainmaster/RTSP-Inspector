package rtsp

import (
	"bufio"
	"net"
	"net/textproto"
	"rtsp-inspector/auth"
)

type Client struct {
	conn        net.Conn
	reader      *bufio.Reader
	tp          *textproto.Reader
	csec        int
	digestAuth  auth.DigestAuth
	Credentials Credentials
}
