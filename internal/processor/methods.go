package processor

import (
	"context"
	"encoding/binary"
	"io"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

type Processor struct {
	client       *rtsp_client.Client
	codecType    types.CodecType
	depacketizer rtp.Depacketizer
	DataChannels DataChannels
}
type DataChannels struct {
	VideoCh     chan []byte
	AudioCh     chan []byte
	RTCPVideoCh chan []rtcp.Packet
	RTCPAudioCh chan []rtcp.Packet
}

func NewProcessor(client *rtsp_client.Client, codecType types.CodecType) *Processor {
	var depacketizer rtp.Depacketizer
	switch codecType {
	case types.H264:
		depacketizer = &codecs.H264Packet{}
	case types.H265:
		depacketizer = &codecs.H265Depacketizer{}
	}

	return &Processor{
		client:    client,
		codecType: codecType,
		DataChannels: DataChannels{
			VideoCh:     make(chan []byte, 100),
			AudioCh:     make(chan []byte, 100),
			RTCPVideoCh: make(chan []rtcp.Packet, 10),
			RTCPAudioCh: make(chan []rtcp.Packet, 10),
		},
		depacketizer: depacketizer,
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
		channel := types.RTPType(channelByte)

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
		case types.RTPTypeVideo:
			rtpPkt := &rtp.Packet{}
			rtpPkt.Unmarshal(payload)
			nalu, _ := p.depacketizer.Unmarshal(rtpPkt.Payload)

			if err != nil || len(nalu) == 0 {
				return
			}
			p.DataChannels.VideoCh <- payload
		case types.RTCPTypeVideo:
			packet, rtcpErr := rtcp.Unmarshal(payload)
			if rtcpErr != nil {
				continue
			}
			p.DataChannels.RTCPVideoCh <- packet
		case types.RTPTypeAudio:
			p.DataChannels.AudioCh <- payload
		case types.RTCPTypeAudio:
			packet, rtcpErr := rtcp.Unmarshal(payload)
			if rtcpErr != nil {
				continue
			}
			p.DataChannels.RTCPAudioCh <- packet
		default:
			continue
		}
	}
}

type FrameInfo struct {
	Codec   string
	Type    string
	IsKey   bool
	NaluRaw int
}

func (p *Processor) GetStats(handler rtp.Depacketizer, payload []byte) *FrameInfo {
	nalu, err := handler.Unmarshal(payload)
	if err != nil || len(nalu) == 0 {
		return nil
	}

	stats := &FrameInfo{}

	if codecType == 264 { // H264
		stats.Codec = "H264"
		// Заголовок H264: [7 битов: 0] [1 бит: тип] -> берем 0x1F
		naluType := nalu[0] & 0x1F
		stats.NaluRaw = int(naluType)

		switch naluType {
		case 5:
			stats.Type = "IDR (I-Frame)"
			stats.IsKey = true
		case 1:
			stats.Type = "Non-IDR (P/B-Frame)"
		case 7:
			stats.Type = "SPS"
		case 8:
			stats.Type = "PPS"
		default:
			stats.Type = "Other"
		}
	} else if codecType == 265 { // H265
		stats.Codec = "H265"
		// Заголовок H265: [1 бит: 0] [6 бит: тип] [3 бита: слой] [6 бит: ...]
		// Тип лежит в 1-м байте, сдвинутый на 1 бит вправо
		naluType := (nalu[0] >> 1) & 0x3F
		stats.NaluRaw = int(naluType)

		switch {
		case naluType >= 16 && naluType <= 21:
			stats.Type = "IRAP (Key Frame)"
			stats.IsKey = true
		case naluType <= 9:
			stats.Type = "Trailing (P/B-Frame)"
		case naluType == 32:
			stats.Type = "VPS"
		case naluType == 33:
			stats.Type = "SPS"
		case naluType == 34:
			stats.Type = "PPS"
		default:
			stats.Type = "Other"
		}
	}

	return stats
}
