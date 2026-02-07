package processor

import (
	"rtsp-inspector/internal/types"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

const (
	videoSampleRate = 90000
	maxLate         = 150
)

type VideoProcessor struct {
	sb    *samplebuilder.SampleBuilder
	codec types.CodecType
}

func NewVideoProcessor(codec types.CodecType) *VideoProcessor {
	var depacketizer rtp.Depacketizer
	switch codec {
	case types.H264:
		depacketizer = &codecs.H264Packet{}
	case types.H265:
		depacketizer = &codecs.H265Depacketizer{}
	default:
		return nil
	}

	return &VideoProcessor{
		sb:    samplebuilder.New(maxLate, depacketizer, videoSampleRate),
		codec: codec,
	}
}

func (v *VideoProcessor) Push(payload []byte) error {
	rtpPkt := &rtp.Packet{}
	err := rtpPkt.Unmarshal(payload)
	if err != nil {
		return err
	}

	v.sb.Push(rtpPkt)
	return nil
}

func (v *VideoProcessor) Pop() *media.Sample {
	return v.sb.Pop()
}

func (v *VideoProcessor) GetFrameInfo(frame *media.Sample) *types.FrameInfo {
	if frame == nil || len(frame.Data) < 5 {
		return nil
	}

	info := &types.FrameInfo{Codec: v.codec, NALUs: make([]types.NALUType, 0)}

	var lookup map[int]types.NALUInfo
	var mask byte
	var shift int

	switch v.codec {
	case types.H264:
		lookup = types.H264TypeLookup
		mask = 0x1F
		shift = 0
	case types.H265:
		lookup = types.H265TypeLookup
		mask = 0x3F
		shift = 1
	}

	forEachNALU(frame.Data, func(naluByte byte, isFirst bool) {
		rawType := int((naluByte >> shift) & mask)

		if meta, ok := lookup[rawType]; ok {
			info.NALUs = append(info.NALUs, meta.Type)
			if meta.IsKey {
				info.IsKey = true
			}
		} else {
			info.NALUs = append(info.NALUs, types.NALU_UNKNOWN)
		}
	})

	return info
}

func forEachNALU(data []byte, handler func(naluByte byte, isFirst bool)) {
	first := true
	for i := 0; i < len(data)-3; i++ {
		if data[i] == 0 && data[i+1] == 0 {
			offset := 0
			if data[i+2] == 1 {
				offset = 3 // 00 00 01
			} else if i+3 < len(data) && data[i+2] == 0 && data[i+3] == 1 {
				offset = 4 // 00 00 00 01
			}

			if offset > 0 && i+offset < len(data) {
				handler(data[i+offset], first)
				first = false
				i += offset
			}
		}
	}
}
