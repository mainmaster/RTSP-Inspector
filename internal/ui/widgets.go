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
	urlEntry     *widget.Entry
	BtnConnect   *widget.Button
	rtpForm      *widget.Form
	rtpLabels    map[types.RTPType]*widget.Label
	naluLabels   map[types.NALUType]*widget.Label
	naluForm     *widget.Form
	logAccordion *widget.Accordion // Вместо LogOutput
	logScroll    *container.Scroll
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
		urlEntry:     url,
		logAccordion: accordion,
		logScroll:    container.NewScroll(accordion),
		BtnConnect:   widget.NewButton("CONNECT", nil),
		rtpForm:      statsForm,
		rtpLabels:    make(map[types.RTPType]*widget.Label),
		naluLabels:   make(map[types.NALUType]*widget.Label),
		naluForm:     naluForm,
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
		ui.logAccordion.Append(item)
		ui.logScroll.ScrollToBottom()
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

func (ui *Widgets) GetURL() string {
	return ui.urlEntry.Text
}

func (ui *Widgets) UpdateRTPCounter(counter map[types.RTPType]int) {
	fyne.Do(func() {
		newElementAdded := false

		if len(ui.rtpLabels) == 0 && len(ui.rtpForm.Items) > 0 {
			ui.rtpForm.Items = nil
		}

		for rtp, count := range counter {
			if _, lux := ui.rtpLabels[rtp]; !lux {
				name := types.RTPTypeNames[rtp]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				ui.rtpLabels[rtp] = newLabel
				ui.rtpForm.Append(name, newLabel)
				newElementAdded = true
			}
			ui.rtpLabels[rtp].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			ui.rtpForm.Refresh()
			ui.logScroll.Refresh()
		}
	})
}

func (ui *Widgets) UpdateNALUCounter(counter map[types.NALUType]int) {
	fyne.Do(func() {
		newElementAdded := false

		for nalu, count := range counter {
			if _, lux := ui.naluLabels[nalu]; !lux {
				name := types.NALUNames[nalu]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				ui.naluLabels[nalu] = newLabel
				ui.naluForm.Append(name, newLabel)
				newElementAdded = true
			}
			ui.naluLabels[nalu].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			ui.naluForm.Refresh()
			ui.logScroll.Refresh()
		}
	})
}

func (ui *Widgets) ClearCounters() {
	ui.rtpLabels = make(map[types.RTPType]*widget.Label)
	ui.naluLabels = make(map[types.NALUType]*widget.Label)

	fyne.Do(func() {
		ui.rtpForm.Items = nil
		ui.rtpForm.Refresh()

		ui.naluForm.Items = nil
		ui.naluForm.Refresh()
	})
}
