package ui

import (
	"rtsp-inspector/clients/rtsp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func StartApp() {
	myApp := app.New()
	window := myApp.NewWindow("RTSP Inspector")

	ui := NewUIWidgets()
	client := &rtsp.Client{}
	h := &Handlers{UI: ui, client: client}

	// Привязываем события
	ui.BtnOpen.OnTapped = h.HandleConnect
	//ui.BtnPlay.OnTapped = h.HandlePlay

	// Сборка интерфейса по частям
	content := container.NewBorder(
		makeTopPanel(ui),
		makeBottomPanel(ui),
		nil,
		nil,
		makeCenterContent(ui),
	)

	// Оборачиваем всё в Padded, чтобы контент не лип к рамке окна
	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

func makeTopPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewBorder(
		nil,         // top
		nil,         // bottom
		nil,         // left
		ui.BtnOpen,  // right (кнопка прижмется к правому краю)
		ui.URLEntry, // center (растянется на всю ширину)
	)
}

func makeCenterContent(ui *Widgets) fyne.CanvasObject {
	// 1. Верхняя часть (логи)
	top := container.NewScroll(ui.LogOutput)
	return top
}

func makeBottomPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		ui.BtnDescribe,
		layout.NewSpacer(),
	)
}
