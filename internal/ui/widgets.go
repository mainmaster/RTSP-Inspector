package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Widgets struct {
	URLEntry   *widget.Entry
	LogOutput  *widget.Entry
	BtnOptions *widget.Button
	BtnDesc    *widget.Button
	BtnSetup   *widget.Button
	BtnClear   *widget.Button
}

func NewUIWidgets() *Widgets {
	// Поле ввода URL
	url := widget.NewEntry()
	url.SetPlaceHolder("rtsp://admin:password@ip:554/stream")

	// Лог (многострочный)
	log := widget.NewEntry()
	log.MultiLine = true
	log.Wrapping = fyne.TextWrapWord

	return &Widgets{
		URLEntry:   url,
		LogOutput:  log,
		BtnOptions: widget.NewButton("OPTIONS", nil),
		BtnDesc:    widget.NewButton("DESCRIBE", nil),
		BtnSetup:   widget.NewButton("SETUP", nil),
		BtnClear:   widget.NewButton("Clear Log", nil),
	}
}

func (ui *Widgets) AppendLog(msg string) {
	ui.LogOutput.SetText(ui.LogOutput.Text + "\n" + msg)
	// Можно добавить логику автоскролла здесь, если обернуть в ScrollContainer
}
