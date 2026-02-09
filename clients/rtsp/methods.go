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
	req.AddCSeq()

	if c.conn == nil {
		u, err := url.Parse(req.URL)
		if err != nil {
			return nil, err
		}
		err = c.Connect(u.Host)
		if err != nil {
			return nil, err
		}
	}

	buildReq := req.Build()
	fmt.Printf(buildReq)
	if _, err := c.conn.Write([]byte(buildReq)); err != nil {
		return nil, err
	}

	h, err := c.readHeaders()
	if err != nil {
		return nil, err
	}

	if h.StatusCode == 401 {
		authHeaderVal := h.Header.Get(authHeader)
		req.SetRealm(findParam(authHeaderVal, realmRegEx))
		req.SetNonce(findParam(authHeaderVal, nonceRegEx))

		req.AddCSeq()

		buildReq = req.Build()
		fmt.Printf(buildReq)
		if _, err := c.conn.Write([]byte(buildReq)); err != nil {
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
