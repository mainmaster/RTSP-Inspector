package handlers

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
	"strings"
	"time"
)

func (h *Handlers) HandleConnect(ctx context.Context) {
	rtspURL := h.ui.GetURL()
	if rtspURL == "" {
		return
	}
	h.rtspURL = rtspURL

	go func() {
		err := h.connect(h.rtspURL)
		if err != nil {
			fmt.Println(err)
			return
		}

		res, err := h.rtspFlow(h.rtspURL)
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(1 * time.Second)

		h.rtpReaderFlow(ctx, res)
		go h.tearDownWaiting(ctx, res)
	}()
}

func (h *Handlers) SetCtxCancel(cancel context.CancelFunc) {
	h.cancel = cancel
}

func (h *Handlers) rtpReaderFlow(ctx context.Context, rtspResponse *RTSPFlowResponse) {
	uiTicker := time.NewTicker(200 * time.Millisecond)
	rtspKeepaliveTicker := time.NewTicker(20 * time.Second)

	vp := processor.NewVideoProcessor(rtspResponse.codecs["video"])

	rtpCh := make(chan types.RTPPacket)
	rtspCh := make(chan rtsp_client.RTSPResponse)
	go func() {
		err := h.client.RTPReader(ctx, rtpCh, rtspCh)
		if err != nil {
			h.cancel()
		}
	}()

	go func() {
		defer uiTicker.Stop()
		for {
			select {
			case <-uiTicker.C:
				h.updateRTPCounter()
				h.updateNALUCounter()
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		defer h.cancel()

		for {
			select {
			case rtspPacket, ok := <-rtspCh:
				if !ok {
					return
				}
				h.ui.AddLogEntry(types.MethodOptions, buildOutputString(rtspPacket.Header, rtspPacket.Body), false)
			case rtpPacket, ok := <-rtpCh:
				if !ok {
					return
				}
				h.si.IncrementRTPCounter(&rtpPacket)

				if rtpPacket.Type != types.RTPTypeVideo {
					// only video
					break
				}

				err := vp.Push(rtpPacket.Payload)
				if err != nil {
					return
				}
				for {
					frame := vp.Pop()
					if frame == nil {
						break
					}
					info := vp.GetFrameInfo(frame)
					if info != nil {
						h.si.IncrementNALUCounter(info.NALUs)
					}
				}
			case <-time.After(time.Second * 5):
				return
			case <-rtspKeepaliveTicker.C:
				req, _ := h.client.NewRequest(types.MethodOptions, h.rtspURL)
				h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
				err := h.client.Send(req)
				if err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *Handlers) rtspFlow(rtspURL string) (*RTSPFlowResponse, error) {
	rtspFlowRes := &RTSPFlowResponse{}

	req, err := h.client.NewRequest(types.MethodOptions, rtspURL)
	if err != nil {
		return nil, err
	}

	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)

	req, err = h.client.NewRequest(types.MethodDescribe, rtspURL)
	if err != nil {
		return nil, err
	}
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

	describeRes, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	codecs, err := describeRes.GetCodecs()
	if err != nil {
		return nil, err
	}
	rtspFlowRes.codecs = codecs
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, describeRes.Body), false)

	rtspFlowRes.sessions = make(map[string]struct{})
	for _, t := range describeRes.GetTrackIDs() {
		req, err = h.client.NewRequest(types.MethodSetup, rtspURL)
		if err != nil {
			return nil, err
		}
		req.SetTrackID(t)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
		setupRes, setupErr := h.client.Do(req)
		if setupErr != nil {
			return nil, setupErr
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(setupRes.Header, setupRes.Body), false)
		rtspFlowRes.sessions[setupRes.GetSessionID()] = struct{}{}
	}

	for sessionID, _ := range rtspFlowRes.sessions {
		req, err = h.client.NewRequest(types.MethodPlay, rtspURL)
		if err != nil {
			return nil, err
		}
		req.SetSessionID(sessionID)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

		res, err = h.client.Do(req)
		if err != nil {
			return nil, err
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)
	}
	return rtspFlowRes, nil
}

func (h *Handlers) connect(rtspURL string) error {
	u, err := url.Parse(rtspURL)
	if err != nil {
		return err
	}

	err = h.client.Connect(*u)
	if err != nil {
		return err
	}

	h.ui.UpdateConnectStatus(true)
	h.IsConnected = true

	h.si.ClearCounters()
	h.ui.ClearCounters()

	return nil
}

func (h *Handlers) updateRTPCounter() {
	counter := h.si.GetRTPCounter()
	h.ui.UpdateRTPCounter(counter)
}

func (h *Handlers) updateNALUCounter() {
	counter := h.si.GetNALUCounter()
	h.ui.UpdateNALUCounter(counter)
}

func (h *Handlers) tearDownWaiting(ctx context.Context, rtspRes *RTSPFlowResponse) {
	for {
		select {
		case <-ctx.Done():
			for s, _ := range rtspRes.sessions {
				req, err := h.client.NewRequest(types.MethodTeardown, h.rtspURL)
				if err != nil {
					return
				}
				req.SetSessionID(s)
				h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
				err = h.client.Send(req)
				if err != nil {
					return
				}
			}
			return
		default:
		}
	}
}

func buildOutputString(headers textproto.MIMEHeader, body []byte) string {
	var output strings.Builder
	for k, v := range headers {
		output.WriteString(fmt.Sprintf("%s: %s", k, v[0]))
		output.WriteString("\r\n")
	}
	output.WriteString("\r\n")
	output.WriteString(string(body))
	return output.String()
}
