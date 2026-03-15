package main

import (
	"rtsp-inspector/internal/handlers"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("RTSP Inspector")
	widgets := ui.NewUIWidgets()

	client := rtsp_client.NewClient()
	h := handlers.NewHandlers(client, widgets)

	widgets.BtnConnect.OnTapped = h.HandleConnect

	ui.CreateAndRun(window, widgets)
}
