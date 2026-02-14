package ui

import (
	"fmt"
	"net/url"

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

	u, err := url.Parse(h.UI.URLEntry.Text)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
		return
	}

	err = h.client.Connect(u.Host)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
	}
	h.UI.AppendLog("Connected: " + u.Host)
}

func (h *Handlers) HandleOptions() {
	req, _ := rtsp.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	h.UI.RequestBody.SetText(req.Build())
}

func (h *Handlers) HandleSend() {
	//req.Payload = h.UI.RequestBody.Text
	res, err := h.client.Do(&req)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
		return
	}
	h.UI.AppendLog(fmt.Sprintf("Status: %v", res.Body))
	h.UI.AppendLog("Public: " + res.Header.Get("Public"))
}
