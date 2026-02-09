package rtsp

import (
	"bufio"
	"net"
	"net/textproto"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	tp     *textproto.Reader
}
