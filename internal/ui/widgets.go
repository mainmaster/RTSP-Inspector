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

func (ui *Widgets) GetURL() string {
	return ui.URLEntry.Text
}

func (ui *Widgets) UpdateRTPCounter(counter map[types.RTPType]int) {
	fyne.Do(func() {
		newElementAdded := false

		if len(ui.RTPLabels) == 0 && len(ui.RTPForm.Items) > 0 {
			ui.RTPForm.Items = nil
		}

		for rtp, count := range counter {
			if _, lux := ui.RTPLabels[rtp]; !lux {
				name := types.RTPTypeNames[rtp]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				ui.RTPLabels[rtp] = newLabel
				ui.RTPForm.Append(name, newLabel)
				newElementAdded = true
			}
			ui.RTPLabels[rtp].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			ui.RTPForm.Refresh()
			ui.LogScroll.Refresh()
		}
	})
}

func (ui *Widgets) UpdateNALUCounter(counter map[types.NALUType]int) {
	fyne.Do(func() {
		newElementAdded := false

		for nalu, count := range counter {
			if _, lux := ui.NALULabels[nalu]; !lux {
				name := types.NALUNames[nalu]
				if name == "" {
					continue
				}

				newLabel := widget.NewLabel(fmt.Sprintf("%d", count))
				ui.NALULabels[nalu] = newLabel
				ui.NALUForm.Append(name, newLabel)
				newElementAdded = true
			}
			ui.NALULabels[nalu].SetText(fmt.Sprintf("%d", count))
		}
		if newElementAdded {
			ui.NALUForm.Refresh()
			ui.LogScroll.Refresh()
		}
	})
}

func (ui *Widgets) ClearCounters() {
	ui.RTPLabels = make(map[types.RTPType]*widget.Label)
	ui.NALULabels = make(map[types.NALUType]*widget.Label)

	fyne.Do(func() {
		ui.RTPForm.Items = nil
		ui.RTPForm.Refresh()

		ui.NALUForm.Items = nil
		ui.NALUForm.Refresh()
	})
}
