package processor

import (
	"context"
	"encoding/binary"
	"io"
	"rtsp-inspector/internal/rtsp_client"

	"github.com/pixelbender/go-sdp/sdp"
)

type Processor struct {
	client       *rtsp_client.Client
	codecType    *sdp.Format
	DataChannels DataChannels
}
type DataChannels struct {
	VideoCh     chan []byte
	AudioCh     chan []byte
	RTCPVideoCh chan []byte
	RTCPAudioCh chan []byte
}

func NewProcessor(client *rtsp_client.Client, codecType *sdp.Format) *Processor {
	return &Processor{
		client:    client,
		codecType: codecType,
		DataChannels: DataChannels{
			VideoCh:     make(chan []byte, 100),
			AudioCh:     make(chan []byte, 100),
			RTCPVideoCh: make(chan []byte, 10),
			RTCPAudioCh: make(chan []byte, 10),
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
			p.DataChannels.RTCPVideoCh <- payload
		case 2:
			p.DataChannels.AudioCh <- payload
		case 3:
			p.DataChannels.RTCPAudioCh <- payload
		default:
			continue
		}
	}
}
