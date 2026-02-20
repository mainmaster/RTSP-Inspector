package ui

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/clients/rtsp"
	"rtsp-inspector/types"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

type Handlers struct {
	UI     *Widgets
	client *rtsp.Client
}

func (h *Handlers) HandleConnect() {
	if h.UI.BtnOpen.Text == "DISCONNECT" {
		err := h.client.Close()
		if err != nil {
			h.UI.AppendLog("Error: " + err.Error())
		}
		h.UI.BtnOpen.SetText("CONNECT")
		return
	}

	if h.UI.URLEntry.Text == "" {
		return
	}

	if !h.client.IsEmptyConnection() {
		_ = h.client.Close()
		h.UI.AppendLog("Close old connection")
	}

	u, err := url.Parse(h.UI.URLEntry.Text)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
		return
	}

	err = h.client.Connect(*u)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
	}
	h.UI.AppendLog("Connected: " + u.Host)
	h.UI.BtnOpen.SetText("DISCONNECT")

	req, _ := h.client.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	h.UI.AppendLog(req.BuildRequest())
	res, _ := h.client.Do(req)

	h.UI.AppendLog(buildOutputString(res.Header, res.Body))

	req, _ = h.client.NewRequest("DESCRIBE", h.UI.URLEntry.Text)
	h.UI.AppendLog(req.BuildRequest())
	res, _ = h.client.Do(req)
	h.UI.AppendLog(buildOutputString(res.Header, res.Body))

	sessionIDs := make(map[string]struct{})
	for _, t := range res.GetTrackIDs() {
		req, _ = h.client.NewRequest("SETUP", h.UI.URLEntry.Text)
		req.SetTrackID(t)
		res, err = h.client.Do(req)
		h.UI.AppendLog(buildOutputString(res.Header, res.Body))
		sessionIDs[res.GetSessionID()] = struct{}{}
	}

	for sessionID, _ := range sessionIDs {
		req, _ = h.client.NewRequest("PLAY", h.UI.URLEntry.Text)
		req.SetSessionID(sessionID)
		h.UI.AppendLog(req.BuildRequest())

		res, err = h.client.Do(req)
		if err != nil {
			h.UI.AppendLog("Error: " + err.Error())
			h.client.Close()
			return
		}
		h.UI.AppendLog(buildOutputString(res.Header, res.Body))
	}

	channels := types.DataChannels{
		VideoCh:     make(chan []byte, 100),
		AudioCh:     make(chan []byte, 100),
		RTCPVideoCh: make(chan []byte, 10),
		RTCPAudioCh: make(chan []byte, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// wait PLAY rtsp response
	time.Sleep(1 * time.Second)

	go h.client.ProcessStream(ctx, channels)

	go func() {
		counter := types.PacketCounter{}

		uiTicker := time.NewTicker(200 * time.Millisecond)
		defer uiTicker.Stop()

		for {
			select {
			case <-channels.VideoCh:
				counter.Video++
			case <-channels.AudioCh:
				counter.Audio++
			case <-channels.RTCPVideoCh:
				counter.RTCPVideo++
				//h.UI.AppendLog(string(rtcpVideo))
			case <-channels.RTCPAudioCh:
				counter.RTCPAudio++
				//h.UI.AppendLog(string(rtcpAudio))
			case <-time.After(time.Second * 5):
				fmt.Println("Тишина в эфире более 5 секунд...")
				return
			case <-uiTicker.C:
				fyne.Do(func() {
					h.UpdateCounter(counter)
				})
			}
		}
	}()

	/*
		Контекст умирает сразу!!!
	*/

}

func (h *Handlers) UpdateCounter(counter types.PacketCounter) {
	h.UI.InfoLabels["Video"].SetText(fmt.Sprintf("%d", counter.Video))
	h.UI.InfoLabels["Audio"].SetText(fmt.Sprintf("%d", counter.Audio))
	h.UI.InfoLabels["RTCPVideo"].SetText(fmt.Sprintf("%d", counter.RTCPVideo))
	h.UI.InfoLabels["RTCPAudio"].SetText(fmt.Sprintf("%d", counter.RTCPAudio))
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
