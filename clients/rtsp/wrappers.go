package rtsp

import "rtsp-inspector/types"

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
