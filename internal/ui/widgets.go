package ui

import (
	"fmt"
	"rtsp-inspector/internal/types"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	connectText    = "CONNECT"
	disconnectText = "DISCONNECT"
)

type Widgets struct {
	URLEntry     *widget.Entry
	BtnConnect   *widget.Button
	RTPForm      *widget.Form
	RTPLabels    map[types.RTPType]*widget.Label
	NALULabels   map[types.NALUType]*widget.Label
	NALUForm     *widget.Form
	LogAccordion *widget.Accordion // Вместо LogOutput
	LogScroll    *container.Scroll
}

func NewUIWidgets() *Widgets {
	url := widget.NewEntry()
	url.SetText("rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1")

	log := widget.NewEntry()
	log.MultiLine = true
	log.MultiLine = true
	log.Wrapping = fyne.TextWrapOff

	requestBody := widget.NewEntry()
	requestBody.MultiLine = true
	log.MultiLine = true
	log.Wrapping = fyne.TextWrapOff

	statsForm := widget.NewForm()
	naluForm := widget.NewForm()

	accordion := widget.NewAccordion()
	accordion.MultiOpen = true

	return &Widgets{
		URLEntry:     url,
		LogAccordion: accordion,
		LogScroll:    container.NewScroll(accordion),
		BtnConnect:   widget.NewButton("CONNECT", nil),
		RTPForm:      statsForm,
		RTPLabels:    make(map[types.RTPType]*widget.Label),
		NALULabels:   make(map[types.NALUType]*widget.Label),
		NALUForm:     naluForm,
	}
}

func (ui *Widgets) AddLogEntry(title types.RTSPMethod, body string, isRequest bool) {
	content := widget.NewMultiLineEntry()
	content.SetText(body)
	content.Wrapping = fyne.TextWrapBreak
	content.TextStyle = fyne.TextStyle{Monospace: true}

	prefix := "▶ [RECV]"
	if isRequest {
		prefix = "◀ [SENT]"
	}
	timestamp := time.Now().Format("15:04:05.000")
	fullTitle := fmt.Sprintf("[%s] %s %s", timestamp, prefix, string(title))
	item := widget.NewAccordionItem(fullTitle, content)

	fyne.Do(func() {
		ui.LogAccordion.Append(item)
		ui.LogScroll.ScrollToBottom()
	})
}

func (ui *Widgets) UpdateConnectStatus(isConnected bool) {
	fyne.Do(func() {
		if isConnected {
			ui.BtnConnect.SetText(disconnectText)
			ui.BtnConnect.Importance = widget.DangerImportance
		} else {
			ui.BtnConnect.SetText(connectText)
			ui.BtnConnect.Importance = widget.HighImportance
		}
		ui.BtnConnect.Refresh()
	})
}
