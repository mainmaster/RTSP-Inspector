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
	rtspURL := h.ui.URLEntry.Text
	if rtspURL == "" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	if !h.isConnected {
		h.connect(rtspURL)
	} else {
		h.disconnect()
	}

	err := h.rtspFlow(rtspURL)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(1 * time.Second)

	h.rtpReaderFlow(ctx)
}

func (h *Handlers) rtpReaderFlow(ctx context.Context) {
	rtpCh := make(chan types.RTPPacket)
	go h.client.RTPReader(ctx, rtpCh)

	counter := &PacketCounter{}

	uiTicker := time.NewTicker(200 * time.Millisecond)

	vp := processor.NewVideoProcessor(h.codecs["video"])
	go func() {
		defer uiTicker.Stop()
		for {
			select {
			case rtpPacket, ok := <-rtpCh:
				if !ok {
					h.cancel()
					return
				}
				h.IncrementCounter(rtpPacket, counter)
				err := vp.Push(rtpPacket.Payload)
				if err != nil {
					// error
					h.cancel()
				}
				frame := vp.Pop()
				if frame == nil {
					break
				}
				info := vp.GetFrameInfo(frame)
				if info != nil {
					fmt.Println(info.NALUs)
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

func (h *Handlers) rtspFlow(rtspURL string) error {
	req, err := h.client.NewRequest("OPTIONS", rtspURL)
	if err != nil {
		return err
	}

	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
	res, err := h.client.Do(req)
	if err != nil {
		return err
	}
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)

	req, err = h.client.NewRequest("DESCRIBE", rtspURL)
	if err != nil {
		return err
	}
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

	describeRes, err := h.client.Do(req)
	if err != nil {
		return err
	}
	codesc, err := describeRes.GetCodecs()
	if err != nil {
		return err
	}
	h.codecs = codesc
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, describeRes.Body), false)

	sessionIDs := make(map[string]struct{})
	for _, t := range describeRes.GetTrackIDs() {
		req, err = h.client.NewRequest("SETUP", rtspURL)
		if err != nil {
			return err
		}
		req.SetTrackID(t)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
		setupRes, setupErr := h.client.Do(req)
		if setupErr != nil {
			return setupErr
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(setupRes.Header, setupRes.Body), false)
		sessionIDs[res.GetSessionID()] = struct{}{}
	}

	for sessionID, _ := range sessionIDs {
		req, err = h.client.NewRequest("PLAY", rtspURL)
		if err != nil {
			return err
		}
		req.SetSessionID(sessionID)
		h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

		res, err = h.client.Do(req)
		if err != nil {
			return err
		}
		h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)
	}
	return nil
}

func (h *Handlers) disconnect() {
	err := h.client.Close()
	if h.cancel != nil {
		h.cancel()
	}
	if err != nil {
		// Ошибка
	}
	h.ui.BtnOpen.SetText("CONNECT")

	if !h.client.IsEmptyConnection() {
		_ = h.client.Close()
		// Ошибка
	}
	h.isConnected = false
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
	h.ui.BtnOpen.SetText("DISCONNECT")
	h.isConnected = true
	h.UpdateCounter(&PacketCounter{})
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
