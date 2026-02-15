package ui

import (
	"fmt"
	"net/url"
	"strings"

	"rtsp-inspector/clients/rtsp"
)

type Handlers struct {
	UI     *Widgets
	client *rtsp.Client
}

func (h *Handlers) HandleConnect() {
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
	h.UI.BtnConnect.Text = "RECONNECT"
}

func (h *Handlers) HandleOptions() {
	req, _ := rtsp.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	payload := h.client.GetPreparedPayload(*req)
	h.UI.RequestBody.SetText(payload.Build())
}

func (h *Handlers) HandleDescribe() {
	req, _ := rtsp.NewRequest("DESCRIBE", h.UI.URLEntry.Text)
	payload := h.client.GetPreparedPayload(*req)
	h.UI.RequestBody.SetText(payload.Build())
}

func (h *Handlers) HandleSetup() {
	req, _ := rtsp.NewRequest("SETUP", h.UI.URLEntry.Text)
	req.TrackID = 1
	payload := h.client.GetPreparedPayload(*req)
	h.UI.RequestBody.SetText(payload.Build())
}

func (h *Handlers) HandlePlay() {
	req, _ := rtsp.NewRequest("PLAY", h.UI.URLEntry.Text)
	payload := h.client.GetPreparedPayload(*req)
	h.UI.RequestBody.SetText(payload.Build())
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
