package ui

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/processor"
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
			h.ui.AppendLog("Error: " + err.Error())
		}
		h.ui.BtnOpen.SetText("CONNECT")
		return
	}

	if h.ui.URLEntry.Text == "" {
		return
	}

	if !h.client.IsEmptyConnection() {
		_ = h.client.Close()
		h.ui.AppendLog("Close old connection")
	}

	u, err := url.Parse(h.ui.URLEntry.Text)
	if err != nil {
		h.ui.AppendLog("Error: " + err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	err = h.client.Connect(*u)
	if err != nil {
		h.ui.AppendLog("Error: " + err.Error())
	}
	h.ui.AppendLog("Connected: " + u.Host)
	h.ui.BtnOpen.SetText("DISCONNECT")

	req, _ := h.client.NewRequest("OPTIONS", h.ui.URLEntry.Text)
	h.ui.AppendLog(req.BuildRequest())
	res, _ := h.client.Do(req)

	h.ui.AppendLog(buildOutputString(res.Header, res.Body))

	req, _ = h.client.NewRequest("DESCRIBE", h.ui.URLEntry.Text)
	h.ui.AppendLog(req.BuildRequest())
	res, _ = h.client.Do(req)
	h.ui.AppendLog(buildOutputString(res.Header, res.Body))

	sessionIDs := make(map[string]struct{})
	for _, t := range res.GetTrackIDs() {
		req, _ = h.client.NewRequest("SETUP", h.ui.URLEntry.Text)
		req.SetTrackID(t)
		res, err = h.client.Do(req)
		h.ui.AppendLog(buildOutputString(res.Header, res.Body))
		sessionIDs[res.GetSessionID()] = struct{}{}
	}

	for sessionID, _ := range sessionIDs {
		req, _ = h.client.NewRequest("PLAY", h.ui.URLEntry.Text)
		req.SetSessionID(sessionID)
		h.ui.AppendLog(req.BuildRequest())

		res, err = h.client.Do(req)
		if err != nil {
			h.ui.AppendLog("Error: " + err.Error())
			h.client.Close()
			return
		}
		h.ui.AppendLog(buildOutputString(res.Header, res.Body))
	}

	// wait PLAY rtsp_client response
	time.Sleep(1 * time.Second)

	p := processor.NewProcessor(h.client)
	go p.StartReadStream(ctx)

	go func() {
		counter := PacketCounter{}

		uiTicker := time.NewTicker(200 * time.Millisecond)
		defer uiTicker.Stop()

		for {
			select {
			case <-p.DataChannels.VideoCh:
				counter.Video++
			case <-p.DataChannels.AudioCh:
				counter.Audio++
			case <-p.DataChannels.RTCPVideoCh:
				counter.RTCPVideo++
			case <-p.DataChannels.RTCPAudioCh:
				counter.RTCPAudio++
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

func (h *Handlers) UpdateCounter(counter PacketCounter) {
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
