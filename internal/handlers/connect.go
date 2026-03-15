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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (h *Handlers) HandleConnect() {
	if h.ui.URLEntry.Text == "" {
		return
	}
	h.rtspURL = h.ui.URLEntry.Text

	if h.ctx.Err() != nil {
		h.ctx, h.cancel = context.WithCancel(context.Background())
	}

	go func() {
		if !h.isConnected {
			h.connect(h.rtspURL)
		} else {
			h.disconnect()
			return
		}

		res, err := h.rtspFlow(h.rtspURL)
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(1 * time.Second)

		h.rtpReaderFlow(h.ctx, res)
		go h.tearDownWaiting(h.ctx, res)
	}()
}

func (h *Handlers) rtpReaderFlow(ctx context.Context, rtspResponse *RTSPFlowResponse) {
	uiTicker := time.NewTicker(200 * time.Millisecond)
	rtspKeepaliveTicker := time.NewTicker(10 * time.Second)
	keyFrameTicker := time.NewTicker(2 * time.Second)

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

		var videoSSRC uint32
		var videoChannel byte

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
				if videoSSRC == 0 && rtpPacket.Type == types.RTPTypeVideo {
					videoSSRC = rtpPacket.GetSSRC()
					videoChannel = rtpPacket.Channel
				}
				h.si.IncrementRTPCounter(&rtpPacket)

				if rtpPacket.Type != types.RTPTypeVideo {
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
			case <-keyFrameTicker.C:
				err := h.client.SendPLI(videoChannel, videoSSRC)
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

func (h *Handlers) connect(rtspURL string) {
	u, err := url.Parse(rtspURL)
	if err != nil {
		// Ошибка
		return
	}

	err = h.client.Connect(*u)
	if err != nil {
		// Ошибка
	}

	h.ui.UpdateConnectStatus(true)
	h.isConnected = true

	h.si.Clear()

	h.ui.RTPLabels = make(map[types.RTPType]*widget.Label)
	h.ui.NALULabels = make(map[types.NALUType]*widget.Label)

	fyne.Do(func() {
		h.ui.RTPForm.Items = nil
		h.ui.RTPForm.Refresh()

		h.ui.NALUForm.Items = nil
		h.ui.NALUForm.Refresh()
	})
}

func (h *Handlers) updateRTPCounter() {
	counter := h.si.GetRTPCounter()

	fyne.Do(func() {
		newElementAdded := false

		if len(h.ui.RTPLabels) == 0 && len(h.ui.RTPForm.Items) > 0 {
			h.ui.RTPForm.Items = nil
		}

		for rtp, count := range counter {
			if _, lux := h.ui.RTPLabels[rtp]; !lux {
				name := types.RTPTypeNames[rtp]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				h.ui.RTPLabels[rtp] = newLabel
				h.ui.RTPForm.Append(name, newLabel)
				newElementAdded = true
			}
			h.ui.RTPLabels[rtp].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			h.ui.RTPForm.Refresh()
			h.ui.LogScroll.Refresh()
		}
	})
}

func (h *Handlers) updateNALUCounter() {
	counter := h.si.GetNALUCounter()

	fyne.Do(func() {
		newElementAdded := false

		for nalu, count := range counter {
			if _, lux := h.ui.NALULabels[nalu]; !lux {
				name := types.NALUNames[nalu]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				h.ui.NALULabels[nalu] = newLabel
				h.ui.NALUForm.Append(name, newLabel)
				newElementAdded = true
			}
			h.ui.NALULabels[nalu].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			h.ui.NALUForm.Refresh()
			h.ui.LogScroll.Refresh()
		}
	})
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
