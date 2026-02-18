package ui

import (
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/clients/rtsp"
	"rtsp-inspector/types"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/pixelbender/go-sdp/sdp"
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
	h.UI.BtnOpen.Text = "DISCONNECT"

	req, _ := h.client.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	h.UI.AppendLog(h.client.BuildRequest(*req))
	res, _ := h.client.Do(req)

	h.UI.AppendLog(buildOutputString(res.Header, res.Body))

	req, _ = h.client.NewRequest("DESCRIBE", h.UI.URLEntry.Text)
	h.UI.AppendLog(h.client.BuildRequest(*req))
	res, _ = h.client.Do(req)
	h.UI.AppendLog(buildOutputString(res.Header, res.Body))

	sdpSess, _ := sdp.ParseString(string(res.Body))

	sessionIDs := make(map[string]struct{})
	for _, track := range sdpSess.Media {
		url := h.UI.URLEntry.Text + "/" + track.Attributes[0].Value
		req, _ = h.client.NewRequest("SETUP", url)
		h.UI.AppendLog(h.client.BuildRequest(*req))
		res, _ = h.client.Do(req)
		h.UI.AppendLog(buildOutputString(res.Header, res.Body))
		sessionIDs[res.Header.Get("Session")] = struct{}{}
	}

	for sessionID, _ := range sessionIDs {
		s := strings.Split(sessionID, ";")[0]
		req, _ = h.client.NewRequest("PLAY", h.UI.URLEntry.Text)
		req.Header.Add("Session", s)
		h.UI.AppendLog(h.client.BuildRequest(*req))
		res, _ = h.client.Do(req)
		h.UI.AppendLog(buildOutputString(res.Header, res.Body))
		h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
	}

	channels := types.DataChannels{
		VideoCh: make(chan []byte, 100),
		AudioCh: make(chan []byte, 100),
		RTCPCh:  make(chan []byte, 10),
		ErrCh:   make(chan error, 1),
	}

	go func() {
		go h.client.ProcessStream(channels)

		counter := types.PacketCounter{}

		uiTicker := time.NewTicker(200 * time.Millisecond)
		defer uiTicker.Stop()

		for {
			select {
			case <-channels.VideoCh:
				counter.Video++
			case <-channels.AudioCh:
				counter.Audio++
			case rtcp := <-channels.RTCPCh:
				counter.RTCP++
				h.UI.AppendLog(string(rtcp))

			case <-time.After(time.Second * 5):
				fmt.Println("Тишина в эфире более 5 секунд...")
				return
			case <-channels.ErrCh:
				return
			case <-uiTicker.C:
				fyne.Do(func() {
					h.UpdateCounter(counter)
				})
			}
		}
	}()
}

func (h *Handlers) UpdateCounter(counter types.PacketCounter) {
	h.UI.InfoLabels["Video"].SetText(fmt.Sprintf("%d", counter.Video))
	h.UI.InfoLabels["Audio"].SetText(fmt.Sprintf("%d", counter.Audio))
	h.UI.InfoLabels["Audio"].SetText(fmt.Sprintf("%d", counter.Audio))
	h.UI.InfoLabels["RTCP"].SetText(fmt.Sprintf("%d", counter.RTCP))
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
