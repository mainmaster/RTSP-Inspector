package rtsp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"rtsp-inspector/auth"
	"rtsp-inspector/internal/common_errors"
	"rtsp-inspector/types"
	"strconv"
)

func (c *Client) Do(req *Request) (*types.Response, error) {
	if c.conn == nil {
		return nil, common_errors.ErrNoConnection
	}

	return c.Send(c.BuildRequest(*req))
}

func (c *Client) Send(payload string) (*types.Response, error) {
	defer func() {
		c.csec++
	}()

	if _, err := c.conn.Write([]byte(payload)); err != nil {
		return nil, err
	}

	h, err := c.readHeaders()
	if err != nil {
		return nil, err
	}

	body, err := c.readBody(h.Header)
	if err != nil {
		return nil, err
	}

	return &types.Response{
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

func (c *Client) readHeaders() (types.Headers, error) {
	line, err := c.tp.ReadLine()
	if err != nil {
		return types.Headers{}, err
	}

	var code int
	fmt.Sscanf(line, "RTSP/1.0 %d", &code)

	headers, err := c.tp.ReadMIMEHeader()
	if err != nil {
		return types.Headers{}, err
	}

	return types.Headers{
		Header:     headers,
		StatusLine: line,
		StatusCode: code,
	}, nil
}

func (c *Client) readBody(headers textproto.MIMEHeader) ([]byte, error) {
	contentLengthStr := headers.Get("Content-Length")
	if contentLengthStr == "" {
		return nil, nil
	}

	size, err := strconv.Atoi(contentLengthStr)
	if err != nil || size <= 0 {
		return nil, nil
	}

	body := make([]byte, size)
	_, err = io.ReadFull(c.reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) Connect(u url.URL) error {
	c.csec = 1
	c.digestAuth = auth.DigestAuth{}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.tp = textproto.NewReader(c.reader)

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

func (c *Client) SetSessionID(sessionID string) {
	c.sessionID = sessionID
}

func (c *Client) GetSessionID() string {
	return c.sessionID
}

func (c *Client) ProcessStream(dc types.DataChannels) {
	defer func() {
		close(dc.VideoCh)
		close(dc.RTCPCh)
		close(dc.AudioCh)
		close(dc.ErrCh)
	}()

	for {
		peek, err := c.reader.Peek(1)
		if err != nil {
			return
		}

		if peek[0] != '$' {
			continue
		}

		c.reader.Discard(1)

		channelByte, _ := c.reader.ReadByte()
		channel := int(channelByte)

		lenBuf := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, lenBuf); err != nil {
			return
		}
		length := binary.BigEndian.Uint16(lenBuf)

		payload := make([]byte, length)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return
		}

		switch channel {
		case 0:
			dc.VideoCh <- payload
		case 1:
			dc.RTCPCh <- payload
		case 2:
			dc.AudioCh <- payload
		case 3:
			dc.RTCPCh <- payload
		default:
			continue

		}
	}
}
