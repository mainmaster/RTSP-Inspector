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
	IsConnected bool
	si          *processor.StreamInspector
	cancel      context.CancelCauseFunc
}

func NewHandlers(client *rtsp_client.Client, widgets *ui.Widgets) *Handlers {

	return &Handlers{
		ui:     widgets,
		client: client,
		si:     processor.NewStreamInspector(),
	}
}

type RTSPFlowResponse struct {
	codecs      map[types.TrackType]types.CodecType
	sessions    map[string]struct{}
	Interleaved []types.Interleaved
}
