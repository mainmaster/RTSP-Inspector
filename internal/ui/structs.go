package ui

import (
	"context"
	"rtsp-inspector/internal/processor"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
)

type Handlers struct {
	rtspURL     string
	sessions    map[string]struct{}
	ui          *Widgets
	client      *rtsp_client.Client
	cancel      context.CancelFunc
	isConnected bool
	codecs      map[string]types.CodecType
	//ssrc        map[int]int
	si *processor.StreamInspector
}
