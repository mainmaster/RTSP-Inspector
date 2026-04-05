package handlers

import (
	"context"
	"fmt"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
)

type Handlers struct {
	rtspURL     string
	updater     UIUpdater
	client      *rtsp_client.Client
	IsConnected bool
	useUDP      bool
	si          *processor.StreamInspector
	cancel      context.CancelCauseFunc
}

func NewHandlers(client *rtsp_client.Client, updater UIUpdater) *Handlers {

	return &Handlers{
		updater: updater,
		client:  client,
		si:      processor.NewStreamInspector(),
	}
}

func (h *Handlers) handleError(err error, context string) {
	if err == nil {
		return
	}
	wrappedErr := fmt.Errorf("%s: %w", context, err)
	h.updater.ShowError(wrappedErr)
}

type RTSPFlowResponse struct {
	codecs      map[types.TrackType]types.CodecType
	sessions    map[string]struct{}
	Interleaved []types.Interleaved
}
