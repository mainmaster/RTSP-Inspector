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
}

func (h *Handlers) HandleOptions() {
	req, _ := rtsp.NewRequest("OPTIONS", h.UI.URLEntry.Text)
	payload := h.client.GetPreparedPayload(*req)
	h.UI.RequestBody.SetText(payload.Build())
}

func (h *Handlers) HandleSend() {
	res, err := h.client.Send(h.UI.RequestBody.Text)
	if err != nil {
		h.UI.AppendLog("Error: " + err.Error())
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
