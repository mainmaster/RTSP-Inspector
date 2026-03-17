package main

import (
	"context"
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

	widgets.BtnConnect.OnTapped = func() {
		var ctx context.Context
		var cancel context.CancelFunc

		if h.IsConnected {
			h.HandelDisconnect(ctx)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
			h.SetCtxCancel(cancel)
			h.HandleConnect(ctx)
		}
		h.HandleConnect(ctx)
	}

	ui.CreateAndRun(window, widgets)
}
