package rtsp_client

import "rtsp-inspector/internal/types"

func (c *Client) Options(rtspURL string) (*RTSPResponse, error) {
	req, err := c.NewRequest(types.MethodOptions, rtspURL)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	res.Method = req.Method
	return res, nil
}

func (c *Client) Describe(rtspURL string) (*RTSPResponse, error) {
	req, err := c.NewRequest(types.MethodDescribe, rtspURL)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	res.Method = req.Method
	return res, nil
}

func (c *Client) Setup(rtspURL string, trackID string) (*RTSPResponse, error) {
	req, err := c.NewRequest(types.MethodSetup, rtspURL)
	if err != nil {
		return nil, err
	}
	req.SetTrackID(trackID)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	res.Method = req.Method
	return res, nil
}

func (c *Client) Play(rtspURL string, sessionID string) (*RTSPResponse, error) {
	req, err := c.NewRequest(types.MethodPlay, rtspURL)
	if err != nil {
		return nil, err
	}
	req.SetSessionID(sessionID)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	res.Method = req.Method
	return res, nil
}

func (c *Client) Teardown(rtspURL string, sessionID string) (*RTSPResponse, error) {
	req, err := c.NewRequest(types.MethodPlay, rtspURL)
	if err != nil {
		return nil, err
	}
	req.SetSessionID(sessionID)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	res.Method = req.Method
	return res, nil
}
