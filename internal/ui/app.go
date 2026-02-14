package ui

import (
	"image/color"
	"rtsp-inspector/clients/rtsp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
	ui.BtnConnect.OnTapped = h.HandleConnect
	ui.BtnOptions.OnTapped = h.HandleOptions
	ui.BtnSend.OnTapped = h.HandleSend
	ui.BtnClear.OnTapped = func() { ui.LogOutput.SetText(""); ui.RequestBody.SetText("") }

	// Сборка интерфейса по частям
	content := container.NewBorder(
		makeTopPanel(ui),
		makeBottomPanel(ui),
		makeLeftPanel(ui),
		nil,
		makeCenterContent(ui),
	)

	// Оборачиваем всё в Padded, чтобы контент не лип к рамке окна
	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

func makeTopPanel(ui *Widgets) fyne.CanvasObject {
	connectRow := container.NewHBox(
		ui.BtnConnect,
		layout.NewSpacer(),
	)

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, 15))

	return container.NewVBox(
		ui.URLEntry,
		connectRow,
		spacer,
	)
}

func makeLeftPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewVBox(
		ui.BtnOptions,
		// Сюда удобно добавлять новые кнопки методов (DESCRIBE, SETUP и т.д.)
	)
}

func makeCenterContent(ui *Widgets) fyne.CanvasObject {
	return container.NewBorder(
		nil, nil, nil, nil,
		container.NewGridWithColumns(2,
			container.NewBorder(nil, nil, nil, ui.BtnSend, container.NewScroll(ui.RequestBody)),
			container.NewScroll(ui.LogOutput),
		),
	)
}

func makeBottomPanel(ui *Widgets) fyne.CanvasObject {
	return container.NewHBox(
		layout.NewSpacer(),
		ui.BtnClear,
	)
}
