package processor

import (
	"context"
	"encoding/binary"
	"github.com/pion/rtcp"
	"github.com/pixelbender/go-sdp/sdp"
	"io"
	"rtsp-inspector/internal/rtsp_client"
)

type Processor struct {
	client       *rtsp_client.Client
	codecType    *sdp.Format
	DataChannels DataChannels
}
type DataChannels struct {
	VideoCh     chan []byte
	AudioCh     chan []byte
	RTCPVideoCh chan []rtcp.Packet
	RTCPAudioCh chan []rtcp.Packet
}

func NewProcessor(client *rtsp_client.Client, codecType *sdp.Format) *Processor {
	return &Processor{
		client:    client,
		codecType: codecType,
		DataChannels: DataChannels{
			VideoCh:     make(chan []byte, 100),
			AudioCh:     make(chan []byte, 100),
			RTCPVideoCh: make(chan []rtcp.Packet, 10),
			RTCPAudioCh: make(chan []rtcp.Packet, 10),
		},
	}
}

func (p *Processor) StartReadStream(ctx context.Context) {
	defer func() {
		close(p.DataChannels.VideoCh)
		close(p.DataChannels.RTCPVideoCh)
		close(p.DataChannels.AudioCh)
		close(p.DataChannels.RTCPAudioCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		peek, err := p.client.Reader.Peek(1)
		if err != nil {
			return
		}

		if peek[0] != '$' {
			continue
		}

		p.client.Reader.Discard(1)

		channelByte, _ := p.client.Reader.ReadByte()
		channel := int(channelByte)

		lenBuf := make([]byte, 2)
		if _, err := io.ReadFull(p.client.Reader, lenBuf); err != nil {
			return
		}
		length := binary.BigEndian.Uint16(lenBuf)

		payload := make([]byte, length)
		if _, err := io.ReadFull(p.client.Reader, payload); err != nil {
			return
		}

		switch channel {
		case 0:
			p.DataChannels.VideoCh <- payload
		case 1:
			data, rtcpErr := rtcp.Unmarshal(payload)
			if rtcpErr != nil {
				continue
			}
			p.DataChannels.RTCPVideoCh <- data
		case 2:
			p.DataChannels.AudioCh <- payload
		case 3:
			data, rtcpErr := rtcp.Unmarshal(payload)
			if rtcpErr != nil {
				continue
			}
			p.DataChannels.RTCPAudioCh <- data
		default:
			continue
		}
	}
}
