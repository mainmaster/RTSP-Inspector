package ui

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/types"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func (h *Handlers) HandleConnect() {
	if h.ui.BtnOpen.Text == "DISCONNECT" {
		err := h.client.Close()
		if h.cancel != nil {
			h.cancel()
		}
		if err != nil {
			// Ошибка
		}
		h.ui.BtnOpen.SetText("CONNECT")
		return
	}

	rtspURL := h.ui.URLEntry.Text
	if rtspURL == "" {
		return
	}

	if !h.client.IsEmptyConnection() {
		_ = h.client.Close()
		// Ошибка
	}

	u, err := url.Parse(rtspURL)
	if err != nil {
		// Ошибка
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	err = h.client.Connect(*u)
	if err != nil {
		// Ошибка
	}
	h.ui.BtnOpen.SetText("DISCONNECT")

	req, _ := h.client.NewRequest("OPTIONS", rtspURL)
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
	res, _ := h.client.Do(req)
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)

	req, _ = h.client.NewRequest("DESCRIBE", rtspURL)
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
	describeRes, _ := h.client.Do(req)
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, describeRes.Body), false)

	sessionIDs := make(map[string]struct{})
	for _, t := range describeRes.GetTrackIDs() {
		req, _ = h.client.NewRequest("SETUP", rtspURL)
		req.SetTrackID(t)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
		setupRes, _ := h.client.Do(req)
		h.ui.AddLogEntry(req.Method, buildOutputString(setupRes.Header, setupRes.Body), false)
		sessionIDs[res.GetSessionID()] = struct{}{}
	}

	for sessionID, _ := range sessionIDs {
		req, _ = h.client.NewRequest("PLAY", rtspURL)
		req.SetSessionID(sessionID)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

		res, err = h.client.Do(req)
		if err != nil {
			// Ошибка
			h.client.Close()
			return
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)
	}

	// wait PLAY rtsp_client response
	time.Sleep(1 * time.Second)

	rtpCh := make(chan types.RTPPacket)
	go h.client.RTPReader(ctx, rtpCh)

	codecs, _ := describeRes.GetCodecs()

	counter := &PacketCounter{}

	uiTicker := time.NewTicker(200 * time.Millisecond)

	vp := processor.NewVideoProcessor(codecs["video"])
	go func() {
		defer uiTicker.Stop()
		for {
			select {
			case rtpPacket := <-rtpCh:
				h.IncrementCounter(rtpPacket, counter)
				err = vp.Push(rtpPacket.Payload)
				if err != nil {
					fmt.Println("Error: " + err.Error())
					h.cancel()
				}
				frame := vp.Pop()
				if frame == nil {
					break // Кадр еще не собран полностью
				}
				info := vp.GetFrameInfo(frame)
				if info != nil {
					fmt.Println(info)
				}
			case <-time.After(time.Second * 5):
				h.cancel()
				return
			case <-uiTicker.C:
				fyne.Do(func() {
					h.UpdateCounter(counter)
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *Handlers) IncrementCounter(packet types.RTPPacket, counter *PacketCounter) {
	switch packet.Type {
	case types.RTPTypeAudio:
		counter.Audio++
	case types.RTPTypeVideo:
		counter.Video++
	case types.RTCPTypeAudio:
		counter.RTCPAudio++
	case types.RTCPTypeVideo:
		counter.RTCPVideo++
	}
}

func (h *Handlers) UpdateCounter(counter *PacketCounter) {
	h.ui.InfoLabels["Video"].SetText(fmt.Sprintf("%d", counter.Video))
	h.ui.InfoLabels["Audio"].SetText(fmt.Sprintf("%d", counter.Audio))
	h.ui.InfoLabels["RTCPVideo"].SetText(fmt.Sprintf("%d", counter.RTCPVideo))
	h.ui.InfoLabels["RTCPAudio"].SetText(fmt.Sprintf("%d", counter.RTCPAudio))
	h.ui.InfoLabels["Packets"].SetText(fmt.Sprintf("%d", counter.Video+counter.Audio+counter.RTCPVideo+counter.RTCPAudio))
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
