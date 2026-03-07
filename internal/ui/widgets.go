package ui

import (
	"fmt"
	"rtsp-inspector/internal/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Widgets struct {
	URLEntry     *widget.Entry
	BtnOpen      *widget.Button
	StatsForm    *widget.Form
	InfoLabels   map[string]*widget.Label
	NALULabels   map[types.NALUType]*widget.Label
	LogAccordion *widget.Accordion // Вместо LogOutput
	LogScroll    *container.Scroll
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

	statsForm := widget.NewForm()
	infoLabels := make(map[string]*widget.Label)
	keys := []string{"Packets", "Video", "Audio", "RTCPVideo", "RTCPAudio"}
	for _, k := range keys {
		lbl := widget.NewLabel("-")
		infoLabels[k] = lbl
		statsForm.Append(k, lbl)
	}

	accordion := widget.NewAccordion()
	accordion.MultiOpen = true

	return &Widgets{
		URLEntry:     url,
		LogAccordion: accordion,
		LogScroll:    container.NewScroll(accordion),
		BtnOpen:      widget.NewButton("CONNECT", nil),
		StatsForm:    statsForm,
		InfoLabels:   infoLabels,
	}
}

func (ui *Widgets) AddLogEntry(title string, body string, isRequest bool) {
	content := widget.NewLabel(body)
	content.Wrapping = fyne.TextWrapBreak

	prefix := "▶ [RECV]"
	if isRequest {
		prefix = "◀ [SENT]"
	}
	title = fmt.Sprintf("%s %s", prefix, title)

	item := widget.NewAccordionItem(title, content)

	fyne.Do(func() {
		ui.LogAccordion.Append(item)
		// Авто-скролл вниз при добавлении
		ui.LogScroll.ScrollToBottom()
	})
}
