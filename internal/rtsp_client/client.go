package rtsp_client

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/rtsp_client/auth"
	"rtsp-inspector/internal/types"
	"strconv"
)

type Client struct {
	conn        net.Conn
	Reader      *bufio.Reader
	tp          *textproto.Reader
	csec        int
	sessionID   string
	digestAuth  auth.DigestAuth
	Credentials Credentials
}

func (c *Client) Do(req *Request) (*Response, error) {
	if c.conn == nil {
		return nil, errors.New("no connection")
	}
	return c.Send(req.BuildRequest())
}

func (c *Client) Send(payload string) (*Response, error) {
	defer func() {
		c.csec++
	}()

	if _, err := c.conn.Write([]byte(payload)); err != nil {
		return nil, err
	}

	h, err := c.readRTSPHeaders()
	if err != nil {
		return nil, err
	}

	body, err := c.readRTSPBody(h.Header)
	if err != nil {
		return nil, err
	}

	return &Response{
		Headers: h,
		Body:    body,
	}, nil
}

func (c *Client) setCredentialsFromURL(u url.URL) {
	pass, _ := u.User.Password()
	username := u.User.Username()

	c.digestAuth.Username = username
	c.digestAuth.Password = pass
}

func (c *Client) readRTSPHeaders() (Headers, error) {
	line, err := c.tp.ReadLine()
	if err != nil {
		return Headers{}, err
	}

	var code int
	fmt.Sscanf(line, "RTSP/1.0 %d", &code)

	headers, err := c.tp.ReadMIMEHeader()
	if err != nil {
		return Headers{}, err
	}

	return Headers{
		Header:     headers,
		StatusLine: line,
		StatusCode: code,
	}, nil
}

func (c *Client) readRTSPBody(headers textproto.MIMEHeader) ([]byte, error) {
	contentLengthStr := headers.Get("Content-Length")
	if contentLengthStr == "" {
		return nil, nil
	}

	size, err := strconv.Atoi(contentLengthStr)
	if err != nil || size <= 0 {
		return nil, nil
	}

	body := make([]byte, size)
	_, err = io.ReadFull(c.Reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) RTPReader(ctx context.Context, ch chan types.RTPPacket) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			peek, err := c.Reader.Peek(1)
			if err != nil {
				return
			}
			if peek[0] != '$' {
				return
			}

			c.Reader.Discard(1)

			channelByte, _ := c.Reader.ReadByte()
			channel := types.RTPType(channelByte)

			lenBuf := make([]byte, 2)
			if _, err := io.ReadFull(c.Reader, lenBuf); err != nil {
				return
			}
			length := binary.BigEndian.Uint16(lenBuf)
			payload := make([]byte, length)
			if _, err := io.ReadFull(c.Reader, payload); err != nil {
				return
			}

			ch <- types.RTPPacket{
				Type:    channel,
				Payload: payload,
			}
		}
	}
}

func (c *Client) Connect(u url.URL) error {
	c.csec = 1
	c.digestAuth = auth.DigestAuth{}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return err
	}
	c.conn = conn
	c.Reader = bufio.NewReader(conn)
	c.tp = textproto.NewReader(c.Reader)

	c.setCredentialsFromURL(u)

	if c.digestAuth.Nonce == "" && c.digestAuth.Realm == "" {
		return c.setDigestHeaders(u)
	}
	return nil
}

func (c *Client) setDigestHeaders(u url.URL) error {
	u.User = nil

	req, err := c.NewRequest("OPTIONS", u.String())
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	authHeaderVal := res.Header.Get(authHeader)
	c.digestAuth.Realm = findParam(authHeaderVal, realmRegEx)
	c.digestAuth.Nonce = findParam(authHeaderVal, nonceRegEx)
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SetCredentials(credentials Credentials) {
	c.digestAuth.Username = credentials.Username
	c.digestAuth.Password = credentials.Password
}

func (c *Client) IsEmptyConnection() bool {
	if c.conn == nil {
		return true
	}
	return false
}
