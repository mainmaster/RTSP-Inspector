package rtsp

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"rtsp-inspector/types"
	"strconv"
	"strings"
)

func (c *Client) send(method string) (*types.Response, error) {
	c.cseq++

	req := c.buildRequest(method, c.digestAuth.Nonce != "")
	if _, err := c.conn.Write([]byte(req)); err != nil {
		return nil, err
	}

	resp, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		authHeader := resp.Header.Get("WWW-Authenticate")
		c.digestAuth.Realm = findParam(authHeader, `realm="([^"]+)"`)
		c.digestAuth.Nonce = findParam(authHeader, `nonce="([^"]+)"`)

		c.cseq++
		reqAuth := c.buildRequest(method, true)
		if _, err := c.conn.Write([]byte(reqAuth)); err != nil {
			return nil, err
		}
		return c.readResponse()
	}
	return resp, nil
}

func (c *Client) Options() (*types.Response, error) {
	return c.send("OPTIONS")
}

func (c *Client) Describe(args ...interface{}) {

}

func (c *Client) Setup(args ...interface{}) {

}

func (c *Client) Play(args ...interface{}) {

}

func (c *Client) Teardown(args ...interface{}) {
	c.send("TEARDOWN")
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) buildRequest(method string, useAuth bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s RTSP/1.0", method, c.rtspURL.String()))
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("CSeq: %d", c.cseq))
	b.WriteString("\r\n")

	if useAuth {
		b.WriteString(fmt.Sprintf("Authorization: %s\r\n", c.digestAuth.GetHeader(method)))
		b.WriteString("\r\n")
	}

	b.WriteString("\r\n")
	return b.String()
}

func (c *Client) readResponse() (*types.Response, error) {
	reader := bufio.NewReader(c.conn)
	tp := textproto.NewReader(reader)

	// 1. Читаем статусную строку
	line, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}

	var code int
	fmt.Sscanf(line, "RTSP/1.0 %d", &code)

	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	resp := &types.Response{
		StatusCode: code,
		StatusLine: line,
		Header:     headers,
	}

	contentLengthStr := headers.Get("Content-Length")
	if contentLengthStr == "" {
		return resp, nil
	}

	size, err := strconv.Atoi(contentLengthStr)
	if err == nil && size > 0 {
		body := make([]byte, size)

		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
	}

	return resp, nil
}
