package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Widgets struct {
	URLEntry    *widget.Entry
	LogOutput   *widget.Entry
	RequestBody *widget.Entry
	BtnConnect  *widget.Button
	BtnOptions  *widget.Button
	BtnDescribe *widget.Button
	BtnSetup    *widget.Button
	BtnPlay     *widget.Button
	BtnClear    *widget.Button
	BtnSend     *widget.Button
}

func NewUIWidgets() *Widgets {
	// Поле ввода URL
	url := widget.NewEntry()
	url.SetText("rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1")

	// Лог (многострочный)
	log := widget.NewEntry()
	log.MultiLine = true
	log.MultiLine = true
	log.Wrapping = fyne.TextWrapOff

	requestBody := widget.NewEntry()
	requestBody.MultiLine = true
	log.MultiLine = true
	log.Wrapping = fyne.TextWrapOff

	return &Widgets{
		URLEntry:    url,
		LogOutput:   log,
		RequestBody: requestBody,
		BtnOptions:  widget.NewButton("OPTIONS", nil),
		BtnConnect:  widget.NewButton("CONNECT", nil),
		BtnDescribe: widget.NewButton("DESCRIBE", nil),
		BtnSetup:    widget.NewButton("SETUP", nil),
		BtnPlay:     widget.NewButton("PLAY", nil),
		BtnClear:    widget.NewButton("Clear Log", nil),
		BtnSend:     widget.NewButton("Send", nil),
	}
}

func (ui *Widgets) AppendLog(msg string) {
	ui.LogOutput.SetText(ui.LogOutput.Text + "\n" + msg)
	// Можно добавить логику автоскролла здесь, если обернуть в ScrollContainer
}
