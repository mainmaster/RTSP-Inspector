package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
	"time"

	"github.com/pion/rtcp"
)

const (
	connectionDelay      = 1 * time.Second
	rtpTimeout           = 5 * time.Second
	keepaliveInterval    = 20 * time.Second
	counterUpdateInterval = 200 * time.Millisecond
)

func (h *Handlers) HandleConnect(ctx context.Context) {
	rtspURL := h.updater.GetURL()
	if rtspURL == "" {
		return
	}
	h.rtspURL = rtspURL
	transport := h.updater.GetTransport()
	h.useUDP = transport == "UDP"

	go func() {
		err := h.connect(ctx, h.rtspURL)
		if err != nil {
			h.handleError(err, "connection failed")
			return
		}

		rtspRes, err := h.processRTSPFlow(h.rtspURL)
		if err != nil {
			h.handleError(err, "RTSP flow failed")
			return
		}

		time.Sleep(connectionDelay)

		rtpCh := make(chan types.RTPPacket)
		rtspCh := make(chan rtsp_client.RTSPResponse)

		h.startRTPReader(ctx, rtpCh, rtspCh)
		h.handleDataChannels(ctx, rtpCh, rtspCh, rtspRes)

		go h.updateCounters(ctx)
		go h.errorWaiting(ctx, rtspRes)
	}()
}

func (h *Handlers) processRTSPFlow(rtspURL string) (*RTSPFlowResponse, error) {
	rtspFlowRes := &RTSPFlowResponse{}

	err := h.sendOptions(rtspURL)
	if err != nil {
		return nil, err
	}

	codecs, trackIDs, err := h.sendDescribe(rtspURL)
	if err != nil {
		return nil, err
	}
	rtspFlowRes.codecs = codecs

	sessions, interleaved, err := h.sendSetup(rtspURL, trackIDs)
	if err != nil {
		return nil, err
	}
	rtspFlowRes.sessions = sessions
	rtspFlowRes.Interleaved = interleaved

	err = h.sendPlay(rtspURL, sessions)
	if err != nil {
		return nil, err
	}

	return rtspFlowRes, nil
}

func (h *Handlers) startRTPReader(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan rtsp_client.RTSPResponse) {
	go func() {
		err := h.client.RTPReader(ctx, rtpCh, rtspCh)
		if err != nil {
			h.cancel(err)
		}
	}()
}

func (h *Handlers) handleDataChannels(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan rtsp_client.RTSPResponse, rtspRes *RTSPFlowResponse) {
	go func() {
		err := h.readDataChannels(ctx, rtpCh, rtspCh, rtspRes)
		if err != nil {
			h.cancel(err)
		}
	}()
}

func (h *Handlers) SetCtxCancel(cancel context.CancelCauseFunc) {
	h.cancel = cancel
}

func (h *Handlers) readDataChannels(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan rtsp_client.RTSPResponse, rtspRes *RTSPFlowResponse) error {
	rtspKeepaliveTicker := time.NewTicker(keepaliveInterval)
	vp := processor.NewVideoProcessor(rtspRes.codecs[types.TrackTypeVideo])
	interleavedMap := getMapFromInterleaved(rtspRes)

	for {
		select {
		case rtspPacket, ok := <-rtspCh:
			if !ok {
				return nil
			}
			h.updater.AddLogEntry(types.MethodOptions, buildOutputString(&rtspPacket), false)
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
				decoded := h.decodeRTCP(rtpPacket.Payload)
				h.updater.AddLogEntry(types.MethodRTCPAudio, fmt.Sprintf("RTCP Packet:\n%s\nDecoded:\n%s", hex.Dump(rtpPacket.Payload), decoded), false)
			case types.RTCPTypeVideo:
				decoded := h.decodeRTCP(rtpPacket.Payload)
				h.updater.AddLogEntry(types.MethodRTCPVideo, fmt.Sprintf("RTCP Packet:\n%s\nDecoded:\n%s", hex.Dump(rtpPacket.Payload), decoded), false)
			}
		case <-time.After(rtpTimeout):
			return fmt.Errorf("timed out waiting for RTP packet")
		case <-rtspKeepaliveTicker.C:
			req, _ := h.client.NewRequest(types.MethodOptions, h.rtspURL)
			h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)
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
	uiTicker := time.NewTicker(counterUpdateInterval)
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

func (h *Handlers) sendOptions(rtspURL string) error {
	req, err := h.client.NewRequest(types.MethodOptions, rtspURL)
	if err != nil {
		return err
	}

	h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)
	res, err := h.client.Do(req)
	if err != nil {
		return err
	}
	h.updater.AddLogEntry(req.Method, buildOutputString(res), false)
	return nil
}

func (h *Handlers) sendDescribe(rtspURL string) (map[types.TrackType]types.CodecType, []types.Track, error) {
	req, err := h.client.NewRequest(types.MethodDescribe, rtspURL)
	if err != nil {
		return nil, nil, err
	}
	h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)

	describeRes, err := h.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	h.updater.AddLogEntry(req.Method, buildOutputString(describeRes), false)
	if describeRes.StatusCode != 200 {
		return nil, nil, fmt.Errorf(describeRes.StatusLine)
	}
	codecs, err := describeRes.GetCodecs()
	if err != nil {
		return nil, nil, err
	}
	trackIDs, err := describeRes.GetTrackIDs()
	if err != nil {
		return nil, nil, err
	}
	return codecs, trackIDs, nil
}

func (h *Handlers) sendSetup(rtspURL string, trackIDs []types.Track) (map[string]struct{}, []types.Interleaved, error) {
	sessions := make(map[string]struct{})
	interleaved := make([]types.Interleaved, len(trackIDs))
	for i, t := range trackIDs {
		req, err := h.client.NewRequest(types.MethodSetup, rtspURL)
		if err != nil {
			return nil, nil, err
		}
		req.SetTrackID(t.ID)
		iCh, _ := req.GetInterleavedChannels()
		interleaved[i] = types.Interleaved{
			InterleavedChannels: iCh,
			TrackType:           t.TrackType,
		}
		if h.useUDP {
			req.SetTransport("RTP/AVP;unicast;client_port=5000-5001")
		}
		h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)
		setupRes, setupErr := h.client.Do(req)
		if setupErr != nil {
			return nil, nil, setupErr
		}
		h.updater.AddLogEntry(req.Method, buildOutputString(setupRes), false)
		sessions[setupRes.GetSessionID()] = struct{}{}
	}
	return sessions, interleaved, nil
}

func (h *Handlers) sendPlay(rtspURL string, sessions map[string]struct{}) error {
	for sessionID := range sessions {
		req, err := h.client.NewRequest(types.MethodPlay, rtspURL)
		if err != nil {
			return err
		}
		req.SetSessionID(sessionID)
		h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)

		res, err := h.client.Do(req)
		if err != nil {
			return err
		}
		h.updater.AddLogEntry(req.Method, buildOutputString(res), false)
	}
	return nil
}

func (h *Handlers) connect(ctx context.Context, rtspURL string) error {
	u, err := url.Parse(rtspURL)
	if err != nil {
		return err
	}

	err = h.client.Connect(ctx, *u)
	if err != nil {
		return err
	}

	h.updater.UpdateConnectStatus(true)
	h.IsConnected = true

	h.si.ClearCounters()
	h.updater.ClearCounters()
	h.updater.ClearLogs()

	return nil
}

func (h *Handlers) updateRTPCounter() {
	counter := h.si.GetRTPCounter()
	h.updater.UpdateRTPCounter(counter)
}

func (h *Handlers) updateNALUCounter() {
	counter := h.si.GetNALUCounter()
	h.updater.UpdateNALUCounter(counter)
}

func (h *Handlers) errorWaiting(ctx context.Context, rtspRes *RTSPFlowResponse) {
	<-ctx.Done()
	for s, _ := range rtspRes.sessions {
		req, err := h.client.NewRequest(types.MethodTeardown, h.rtspURL)
		if err != nil {
			return
		}
		req.SetSessionID(s)
		h.updater.AddLogEntry(req.Method, req.BuildRequest(), true)

		err = context.Cause(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			h.handleError(err, "teardown error")
		}

		err = h.client.Send(req)
		if err != nil {
			return
		}
	}
}

func (h *Handlers) decodeRTCP(data []byte) string {
	packets, err := rtcp.Unmarshal(data)
	if err != nil {
		return fmt.Sprintf("Failed to decode RTCP: %v", err)
	}
	var decoded string
	for _, p := range packets {
		switch pkt := p.(type) {
		case *rtcp.ReceiverReport:
			decoded += fmt.Sprintf("ReceiverReport: SSRC=%d, Reports=%d\n", pkt.SSRC, len(pkt.Reports))
		case *rtcp.SenderReport:
			decoded += fmt.Sprintf("SenderReport: SSRC=%d, NTP=%d, RTP=%d\n", pkt.SSRC, pkt.NTPTime, pkt.RTPTime)
		case *rtcp.SourceDescription:
			decoded += fmt.Sprintf("SourceDescription: Chunks=%d\n", len(pkt.Chunks))
		case *rtcp.Goodbye:
			decoded += fmt.Sprintf("Goodbye: Sources=%v\n", pkt.Sources)
		case *rtcp.PictureLossIndication:
			decoded += fmt.Sprintf("PictureLossIndication: MediaSSRC=%d\n", pkt.MediaSSRC)
		case *rtcp.TransportLayerNack:
			decoded += fmt.Sprintf("TransportLayerNack: MediaSSRC=%d, Nacks=%d\n", pkt.MediaSSRC, len(pkt.Nacks))
		default:
			decoded += fmt.Sprintf("Unknown RTCP Packet: %T\n", p)
		}
	}
	return decoded
}
