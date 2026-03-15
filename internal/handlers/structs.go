package handlers

import (
	"context"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
	"rtsp-inspector/internal/ui"
)

type Handlers struct {
	rtspURL     string
	ui          *ui.Widgets
	client      *rtsp_client.Client
	ctx         context.Context
	cancel      context.CancelFunc
	isConnected bool
	si          *processor.StreamInspector
}

func NewHandlers(client *rtsp_client.Client, widgets *ui.Widgets) *Handlers {
	ctx, cancel := context.WithCancel(context.Background())

	return &Handlers{
		ui:     widgets,
		client: client,
		si:     processor.NewStreamInspector(),
		ctx:    ctx,
		cancel: cancel,
	}
}

type RTSPFlowResponse struct {
	codecs   map[string]types.CodecType
	sessions map[string]struct{}
}
