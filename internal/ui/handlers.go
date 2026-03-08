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
	"fyne.io/fyne/v2/widget"
)

func (h *Handlers) HandleConnect() {
	rtspURL := h.ui.URLEntry.Text
	if rtspURL == "" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	go func() {
		if !h.isConnected {
			h.connect(rtspURL)
		} else {
			h.disconnect()
			return
		}

		err := h.rtspFlow(rtspURL)
		if err != nil {
			fmt.Println(err)
		}

		//time.Sleep(1 * time.Second)

		h.rtpReaderFlow(ctx)
	}()
}

func (h *Handlers) rtpReaderFlow(ctx context.Context) {
	rtpCh := make(chan types.RTPPacket)
	go h.client.RTPReader(ctx, rtpCh)

	h.pc = &PacketCounter{}

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
				h.incrementCounter(rtpPacket)
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
					h.incrementNALUCounter(info.NALUs)
				}
			case <-time.After(time.Second * 5):
				h.cancel()
				return
			case <-uiTicker.C:
				go h.updateNTPCounter()
				go h.updateNALUCounter()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *Handlers) rtspFlow(rtspURL string) error {
	req, err := h.client.NewRequest(types.MethodOptions, rtspURL)
	if err != nil {
		return err
	}

	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)
	res, err := h.client.Do(req)
	if err != nil {
		return err
	}
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, res.Body), false)

	req, err = h.client.NewRequest(types.MethodDescribe, rtspURL)
	if err != nil {
		return err
	}
	h.ui.AddLogEntry(req.Method, req.BuildRequest(), true)

	describeRes, err := h.client.Do(req)
	if err != nil {
		return err
	}
	codecs, err := describeRes.GetCodecs()
	if err != nil {
		return err
	}
	h.codecs = codecs
	h.ui.AddLogEntry(req.Method, buildOutputString(res.Header, describeRes.Body), false)

	sessionIDs := make(map[string]struct{})
	for _, t := range describeRes.GetTrackIDs() {
		req, err = h.client.NewRequest(types.MethodSetup, rtspURL)
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
		req, err = h.client.NewRequest(types.MethodPlay, rtspURL)
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

	fyne.Do(func() {
		h.ui.BtnOpen.SetText("CONNECT")
	})

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

	fyne.Do(func() {
		h.ui.BtnOpen.SetText("DISCONNECT")
	})
	h.isConnected = true
	h.pc = &PacketCounter{}
	h.naluCounter = map[types.NALUType]int{}
	go h.updateNTPCounter()
	go h.updateNALUCounter()
}

func (h *Handlers) incrementCounter(packet types.RTPPacket) {
	switch packet.Type {
	case types.RTPTypeAudio:
		h.pc.Audio++
	case types.RTPTypeVideo:
		h.pc.Video++
	case types.RTCPTypeAudio:
		h.pc.RTCPAudio++
	case types.RTCPTypeVideo:
		h.pc.RTCPVideo++
	}
}

func (h *Handlers) updateNTPCounter() {
	fyne.Do(func() {
		h.ui.InfoLabels["Video"].SetText(fmt.Sprintf("%d", h.pc.Video))
		h.ui.InfoLabels["Audio"].SetText(fmt.Sprintf("%d", h.pc.Audio))
		h.ui.InfoLabels["RTCPVideo"].SetText(fmt.Sprintf("%d", h.pc.RTCPVideo))
		h.ui.InfoLabels["RTCPAudio"].SetText(fmt.Sprintf("%d", h.pc.RTCPAudio))
		h.ui.InfoLabels["Packets"].SetText(fmt.Sprintf("%d", h.pc.Video+h.pc.Audio+h.pc.RTCPVideo+h.pc.RTCPAudio))
	})
}

func (h *Handlers) incrementNALUCounter(nalus []types.NALUType) {
	for _, n := range nalus {
		h.naluCounter[n]++
	}
}

func (h *Handlers) updateNALUCounter() {
	for nalu, count := range h.naluCounter {
		if _, lux := h.ui.NALULabels[nalu]; !lux {
			name := types.NALUNames[nalu]
			if name == "" {
				continue
			}

			newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
			h.ui.NALULabels[nalu] = newLabel
			fyne.Do(func() {
				h.ui.NALUForm.Append(name, newLabel)
			})
		}

		// Обновляем значение
		fyne.Do(func() {
			h.ui.NALULabels[nalu].SetText(fmt.Sprintf("%d", count))
		})
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
