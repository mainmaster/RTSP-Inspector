package ui

import (
	"context"
	"rtsp-inspector/internal/rtsp_client"
)

type Handlers struct {
	ui     *Widgets
	client *rtsp_client.Client
	cancel context.CancelFunc
}

type PacketCounter struct {
	Video     int
	Audio     int
	RTCPVideo int
	RTCPAudio int
}
