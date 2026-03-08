package ui

import (
	"context"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
	"sync"
)

type Handlers struct {
	ui          *Widgets
	client      *rtsp_client.Client
	cancel      context.CancelFunc
	isConnected bool
	codecs      map[string]types.CodecType
	pc          *PacketCounter
	naluCounter map[types.NALUType]int
	mu          sync.Mutex
}

type PacketCounter struct {
	Video     int
	Audio     int
	RTCPVideo int
	RTCPAudio int
}
