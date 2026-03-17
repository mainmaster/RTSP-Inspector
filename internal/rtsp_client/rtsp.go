package rtsp_client

import (
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/types"
	"strconv"
)

func (c *Client) Do(req *Request) (*RTSPResponse, error) {
	if c.conn == nil {
		return nil, errors.New("no connection")
	}
	err := c.Send(req)
	if err != nil {
		return nil, err
	}
	return c.readResponse()
}

func (c *Client) Send(req *Request) error {
	defer func() {
		c.csec++
	}()
	if _, err := c.conn.Write([]byte(req.BuildRequest())); err != nil {
		return err
	}
	return nil
}

func (c *Client) readResponse() (*RTSPResponse, error) {
	h, err := c.readRTSPHeaders()
	if err != nil {
		return nil, err
	}

	body, err := c.readRTSPBody(h.Header)
	if err != nil {
		return nil, err
	}

	return &RTSPResponse{
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

func (c *Client) readRTSPHeaders() (types.Headers, error) {
	line, err := c.tp.ReadLine()
	if err != nil {
		return types.Headers{}, err
	}

	var code int
	_, err = fmt.Sscanf(line, "RTSP/1.0 %d", &code)

	if err != nil {
		return types.Headers{}, fmt.Errorf("can not parse RTSP response %s", line)
	}

	if code != 200 {
		return types.Headers{}, fmt.Errorf("invalid RTSP code - %d, %s", code, line)
	}

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
	_, err = io.ReadFull(c.reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
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
