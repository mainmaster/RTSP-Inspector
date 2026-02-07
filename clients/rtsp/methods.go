package rtsp

import (
	"fmt"
	"io"
	"net/textproto"
	"rtsp-inspector/types"
	"strconv"
	"strings"
)

const (
	realmRegEx = `realm="([^"]+)"`
	nonceRegEx = `nonce="([^"]+)"`
	authHeader = "WWW-Authenticate"
)

func (c *Client) do(method types.RTSPMethod) (*types.Response, error) {
	c.cseq++

	req := c.buildRequest(method, c.digestAuth.Nonce != "")
	if _, err := c.conn.Write([]byte(req)); err != nil {
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

		c.cseq++
		reqAuth := c.buildRequest(method, true)
		if _, err := c.conn.Write([]byte(reqAuth)); err != nil {
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

func (c *Client) Options() (*types.Response, error) {
	return c.do(types.MethodOptions)
}

func (c *Client) Describe() (*types.Response, error) {
	return c.do(types.MethodDescribe)
}

func (c *Client) Setup(args ...interface{}) {

}

func (c *Client) Play(args ...interface{}) {

}

func (c *Client) Teardown(args ...interface{}) {
	c.do("TEARDOWN")
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) buildRequest(method types.RTSPMethod, useAuth bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s RTSP/1.0", method, c.rtspURL.String()))
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("CSeq: %d", c.cseq))
	b.WriteString("\r\n")
	if method == types.MethodDescribe {
		b.WriteString("Accept: application/sdp")
		b.WriteString("\r\n")
	}
	if useAuth {
		b.WriteString(fmt.Sprintf("Authorization: %s", c.digestAuth.GetHeader(method)))
		b.WriteString("\r\n")
	}
	b.WriteString("User-Agent: RTSP-Inspector")
	b.WriteString("\r\n")
	b.WriteString("\r\n")
	return b.String()
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
