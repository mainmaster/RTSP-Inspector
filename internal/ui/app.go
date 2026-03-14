package ui

import (
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func StartApp() {
	myApp := app.New()
	window := myApp.NewWindow("RTSP Inspector")

	ui := NewUIWidgets()
	client := rtsp_client.NewClient()
	h := &Handlers{
		ui:       ui,
		client:   client,
		si:       processor.NewStreamInspector(),
		sessions: make(map[string]struct{}),
	}

	ui.BtnOpen.OnTapped = h.HandleConnect

	content := container.NewBorder(
		makeTopPanel(ui),
		makeBottomPanel(),
		nil,
		nil,
		makeCenterContent(ui),
	)

	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

func makeTopPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		nil,
		ui.BtnOpen,
		ui.URLEntry,
	)
}

func makeCenterContent(ui *Widgets) fyne.CanvasObject {
	top := ui.LogScroll
	leftStats := container.NewVScroll(ui.RTPForm)
	rightStats := container.NewVScroll(ui.NALUForm)
	bottomSplit := container.NewHSplit(leftStats, rightStats)
	bottomSplit.Offset = 0.2
	mainSplit := container.NewVSplit(top, bottomSplit)
	mainSplit.Offset = 0.6

	return mainSplit
}

func makeBottomPanel() fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		layout.NewSpacer(),
	)
}
