package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"rtsp-inspector/types"
	"strconv"
)

const (
	realmRegEx = `realm="([^"]+)"`
	nonceRegEx = `nonce="([^"]+)"`
	authHeader = "WWW-Authenticate"
)

func (c *Client) Do(req *Request) (*types.Response, error) {
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}

	if c.conn == nil {
		err = c.Connect(u.Host)
		if err != nil {
			return nil, err
		}
	}

	c.csec++
	req.SetCSeq(c.csec)

	c.setCredentialsFromURL(*u)

	if c.digestAuth.Nonce != "" {
		req.Header.Add("Authorization", c.digestAuth.GetHeader(req.Method, req.URL))
	}

	x := req.Build()
	if _, err := c.conn.Write([]byte(x)); err != nil {
		return nil, err
	}

	h, err := c.readHeaders()
	if err != nil {
		return nil, err
	}

	if h.StatusCode == 401 {
		authHeaderVal := h.Header.Get(authHeader)
		c.digestAuth.Realm = findParam(authHeaderVal, realmRegEx)
		c.digestAuth.Nonce = findParam(authHeaderVal, nonceRegEx)
		req.SetCSeq(c.csec + 1)

		if _, err = c.conn.Write([]byte(req.Build())); err != nil {
			return nil, err
		}

		h, err = c.readHeaders()
		if err != nil {
			return nil, err
		}
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

func (c *Client) Connect(host string) error {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.tp = textproto.NewReader(c.reader)

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) setCredentialsFromURL(u url.URL) {
	pass, _ := u.User.Password()
	username := u.User.Username()

	c.digestAuth.Username = username
	c.digestAuth.Password = pass
}

func (c *Client) SetCredentials(credentials Credentials) {
	c.digestAuth.Username = credentials.Username
	c.digestAuth.Password = credentials.Password
}
