package ui

import (
	"context"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
)

type Handlers struct {
	rtspURL     string
	ui          *Widgets
	client      *rtsp_client.Client
	cancel      context.CancelFunc
	isConnected bool
	si          *processor.StreamInspector
}

type RTSPFlowResponse struct {
	codecs   map[string]types.CodecType
	sessions map[string]struct{}
}
