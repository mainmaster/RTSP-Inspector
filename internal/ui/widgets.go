package ui

import (
	"fmt"
	"rtsp-inspector/internal/types"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const (
	connectText    = "CONNECT"
	disconnectText = "DISCONNECT"
	timeFormat     = "15:04:05.000"
	logFormat      = "[%s] %s %s"
	defaultRTSP    = "rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1"
	recvPrefix     = "▶ [RECV]"
	sentPrefix     = "◀ [SENT]"
)

type LogEntry struct {
	Title string
	Body  string
}

type Widgets struct {
	urlEntry   *widget.Entry
	BtnConnect *widget.Button
	rtpForm    *widget.Form
	rtpLabels  map[types.RTPType]*widget.Label
	naluLabels map[types.NALUType]*widget.Label
	naluForm   *widget.Form
	logs       []LogEntry
	logList    *widget.List
	detailView *widget.Entry // Правая часть: содержимое лога
}

func NewUIWidgets() *Widgets {
	ui := &Widgets{
		urlEntry:   widget.NewEntry(),
		rtpForm:    widget.NewForm(),
		rtpLabels:  make(map[types.RTPType]*widget.Label),
		naluLabels: make(map[types.NALUType]*widget.Label),
		naluForm:   widget.NewForm(),
		BtnConnect: widget.NewButton(connectText, nil),
		logs:       []LogEntry{},
	}
	ui.urlEntry.SetText(defaultRTSP)

	ui.detailView = widget.NewMultiLineEntry()
	ui.detailView.Wrapping = fyne.TextWrapBreak
	ui.detailView.TextStyle = fyne.TextStyle{Monospace: true}

	ui.logList = widget.NewList(
		func() int { return len(ui.logs) },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(ui.logs[id].Title)
		},
	)

	ui.logList.OnSelected = func(id widget.ListItemID) {
		ui.detailView.SetText(ui.logs[id].Body)
	}

	return ui
}
func (ui *Widgets) AddLogEntry(title types.RTSPMethod, body string, isRequest bool) {
	prefix := recvPrefix
	if isRequest {
		prefix = sentPrefix
	}
	fullTitle := fmt.Sprintf(logFormat, time.Now().Format(timeFormat), prefix, string(title))

	fyne.Do(func() {
		ui.logs = append(ui.logs, LogEntry{Title: fullTitle, Body: body})
		ui.logList.Refresh()
		ui.logList.Select(len(ui.logs) - 1)
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
