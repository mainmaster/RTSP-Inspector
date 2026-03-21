package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
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
			h.ui.ShowError(err)
			return
		}

		rtspRes, err := h.rtspFlow(h.rtspURL)
		if err != nil {
			h.ui.ShowError(err)
			return
		}

		time.Sleep(1 * time.Second)

		rtpCh := make(chan types.RTPPacket)
		rtspCh := make(chan rtsp_client.RTSPResponse)

		go func() {
			err = h.client.RTPReader(ctx, rtpCh, rtspCh)
			if err != nil {
				h.cancel(err)
			}
		}()

		go func() {
			err = h.readDataChannels(ctx, rtpCh, rtspCh, rtspRes)
			if err != nil {
				h.cancel(err)
			}
		}()

		go h.updateCounters(ctx)
		go h.errorWaiting(ctx, rtspRes)
	}()
}

func (h *Handlers) SetCtxCancel(cancel context.CancelCauseFunc) {
	h.cancel = cancel
}

func (h *Handlers) readDataChannels(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan rtsp_client.RTSPResponse, rtspRes *RTSPFlowResponse) error {
	rtspKeepaliveTicker := time.NewTicker(20 * time.Second)
	vp := processor.NewVideoProcessor(rtspRes.codecs[types.TrackTypeVideo])
	interleavedMap := getMapFromInterleaved(rtspRes)

	for {
		select {
		case rtspPacket, ok := <-rtspCh:
			if !ok {
				return nil
			}
			h.ui.AddLogEntry(types.MethodOptions, buildOutputString(&rtspPacket), false)
		case rtpPacket, ok := <-rtpCh:
			if !ok {
				return nil
			}

			rtpType := interleavedMap[rtpPacket.Channel]
			h.si.IncrementRTPCounter(rtpType)

			switch rtpType {
			case types.RTPTypeAudio:
				break
			case types.RTPTypeVideo:
				err := vp.Push(rtpPacket.Payload)
				if err != nil {
					return err
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
			case types.RTCPTypeAudio:
				break
			case types.RTCPTypeVideo:
				break
			}
		case <-time.After(time.Second * 5):
			return fmt.Errorf("timed out waiting for RTP packet")
		case <-rtspKeepaliveTicker.C:
			req, _ := h.client.NewRequest(types.MethodOptions, h.rtspURL)
			h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
			err := h.client.Send(req)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *Handlers) updateCounters(ctx context.Context) {
	uiTicker := time.NewTicker(200 * time.Millisecond)
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
	h.ui.AddLogEntry(req.Method, buildOutputString(res), false)

	req, err = h.client.NewRequest(types.MethodDescribe, rtspURL)
	if err != nil {
		return nil, err
	}
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

	describeRes, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	h.ui.AddLogEntry(req.Method, buildOutputString(describeRes), false)
	if describeRes.StatusCode != 200 {
		return nil, fmt.Errorf(describeRes.StatusLine)
	}
	codecs, err := describeRes.GetCodecs()
	if err != nil {
		return nil, err
	}
	rtspFlowRes.codecs = codecs

	rtspFlowRes.sessions = make(map[string]struct{})
	trackIDs, err := describeRes.GetTrackIDs()
	if err != nil {
		return nil, err
	}

	rtspFlowRes.Interleaved = make([]types.Interleaved, len(trackIDs))
	for _, t := range trackIDs {
		req, err = h.client.NewRequest(types.MethodSetup, rtspURL)
		if err != nil {
			return nil, err
		}
		req.SetTrackID(t.ID)
		iCh, _ := req.GetInterleavedChannels()
		rtspFlowRes.Interleaved = append(
			rtspFlowRes.Interleaved,
			types.Interleaved{
				InterleavedChannels: iCh,
				TrackType:           t.TrackType,
			})
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
		setupRes, setupErr := h.client.Do(req)
		if setupErr != nil {
			return nil, setupErr
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(setupRes), false)
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
		h.ui.AddLogEntry(req.Method, buildOutputString(res), false)
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
	h.ui.ClearLogs()

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

func (h *Handlers) errorWaiting(ctx context.Context, rtspRes *RTSPFlowResponse) {
	<-ctx.Done()
	for s, _ := range rtspRes.sessions {
		req, err := h.client.NewRequest(types.MethodTeardown, h.rtspURL)
		if err != nil {
			return
		}
		req.SetSessionID(s)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

		err = context.Cause(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			h.ui.ShowError(err)
		}

		err = h.client.Send(req)
		if err != nil {
			return
		}
	}
}
