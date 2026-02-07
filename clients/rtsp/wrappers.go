package rtsp

import (
	"rtsp-inspector/types"

	"github.com/pixelbender/go-sdp/sdp"
)

func (c *Client) Options() (*types.Response, error) {
	return c.do(types.MethodOptions)
}

func (c *Client) Describe() (*types.Response, error) {
	res, err := c.do(types.MethodDescribe)
	if err != nil {
		return nil, err
	}

	sdpSess, err := sdp.ParseString(string(res.Body))
	if err != nil {
		return nil, err
	}
	res.SDPSession = sdpSess

	return res, err
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
