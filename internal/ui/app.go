package ui

import (
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
	client := &rtsp_client.Client{}
	h := &Handlers{ui: ui, client: client}

	ui.BtnOpen.OnTapped = h.HandleConnect

	content := container.NewBorder(
		makeTopPanel(ui),
		makeBottomPanel(ui),
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
	gridHoriz := container.NewHBox(
		ui.StatsForm,
	)
	gridFixed := container.NewVBox(gridHoriz, layout.NewSpacer())
	bottom := container.NewStack(container.NewScroll(gridFixed))

	split := container.NewVSplit(top, bottom)
	split.Offset = 0.5
	return split
}

func makeBottomPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		layout.NewSpacer(),
	)
}
