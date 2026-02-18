package ui

import (
	"fmt"
	"net/textproto"
	"net/url"
	"rtsp-inspector/clients/rtsp"
	"strings"

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

	for _, track := range sdpSess.Media {
		fmt.Println(track)
		url := h.UI.URLEntry.Text + "/" + track.Attributes[0].Value
		req, _ = h.client.NewRequest("SETUP", url)
		h.UI.AppendLog(h.client.BuildRequest(*req))
		res, _ = h.client.Do(req)
		h.UI.AppendLog(buildOutputString(res.Header, res.Body))
	}

}

func (h *Handlers) HandlePlay() {
	req, _ := h.client.NewRequest("PLAY", h.UI.URLEntry.Text)
	h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
}

/*
func (h *Handlers) HandleOptions() {
	req, _ := h.client.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
}

func (h *Handlers) HandleDescribe() {
	req, _ := h.client.NewRequest("DESCRIBE", h.UI.URLEntry.Text)
	h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
}

func (h *Handlers) HandleSetup() {
	req, _ := h.client.NewRequest("SETUP", h.UI.URLEntry.Text)
	req.TrackID = 1
	h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
}

func (h *Handlers) HandlePlay() {
	req, _ := h.client.NewRequest("PLAY", h.UI.URLEntry.Text)
	h.UI.RequestBody.SetText(h.client.BuildRequest(*req))
}

func (h *Handlers) HandleSend() {
	res, err := h.client.Send(h.UI.RequestBody.Text)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
		h.UI.AppendLog("Reconnect please")
		return
	}
	var output strings.Builder

	output.WriteString(res.StatusLine)
	output.WriteString("\r\n")
	for k, v := range res.Header {
		output.WriteString(fmt.Sprintf("%s: %s", k, v[0]))
		output.WriteString("\r\n")
	}
	output.WriteString("\r\n")
	output.WriteString(string(res.Body))
	h.UI.LogOutput.SetText(output.String())
}

*/

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
