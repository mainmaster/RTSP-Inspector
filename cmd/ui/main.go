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
	widgets := ui.NewUIWidgets(window)

	client := rtsp_client.NewClient()
	h := handlers.NewHandlers(client, widgets)

	var ctx context.Context
	var cancel context.CancelCauseFunc

	widgets.BtnConnect.OnTapped = func() {
		if h.IsConnected {
			h.HandelDisconnect(ctx)
		} else {
			ctx, cancel = context.WithCancelCause(context.Background())
			h.SetCtxCancel(cancel)
			h.HandleConnect(ctx)
		}
	}

	ui.CreateAndRun(window, widgets)
}
