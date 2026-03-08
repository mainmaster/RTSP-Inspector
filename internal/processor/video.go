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
	maxLate         = 100
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

	switch v.codec {
	case types.H264:
		return v.processH264(frame.Data)
	case types.H265:
		return v.processH265(frame.Data)
	default:
		return &types.FrameInfo{Codec: v.codec}
	}
}

func (v *VideoProcessor) processH264(data []byte) *types.FrameInfo {
	info := &types.FrameInfo{Codec: types.H264}

	forEachNALU(data, func(naluByte byte, isFirst bool) {
		naluType := int(naluByte & 0x1F)

		var t types.NALUType
		switch naluType {
		case 1:
			t = types.H264_NALU_NON_IDR
		case 5:
			t = types.H264_NALU_IDR
			info.IsKey = true
		case 6:
			t = types.H264_NALU_SEI
		case 7:
			t = types.H264_NALU_SPS
		case 8:
			t = types.H264_NALU_PPS
		default:
			t = types.NALU_UNKNOWN
		}
		info.NALUs = append(info.NALUs, t)
	})
	return info
}

func (v *VideoProcessor) processH265(data []byte) *types.FrameInfo {
	info := &types.FrameInfo{
		Codec: types.H265,
		NALUs: make([]types.NALUType, 0),
	}

	forEachNALU(data, func(naluByte byte, isFirst bool) {
		naluType := int((naluByte >> 1) & 0x3F)
		var t types.NALUType
		isKeyNALU := false

		switch {
		case naluType == 1:
			t = types.H265_NALU_TRAIL_R
		case naluType == 4:
			t = types.H265_NALU_RASL_R
		case naluType >= 16 && naluType <= 19:
			t = types.H265_NALU_IDR_W
			isKeyNALU = true
		case naluType == 20:
			t = types.H265_NALU_IDR_N
			isKeyNALU = true
		case naluType == 21:
			t = types.H265_NALU_CRA
			isKeyNALU = true
		case naluType == 23:
			t = types.H265_NALU_IDR_W_RADL
			isKeyNALU = true
		case naluType == 24:
			t = types.H265_NALU_IDR_W_RADL
			isKeyNALU = true
		case naluType == 32:
			t = types.H265_NALU_VPS
		case naluType == 33:
			t = types.H265_NALU_SPS
		case naluType == 34:
			t = types.H265_NALU_PPS
		case naluType == 35:
			t = types.H265_NALU_AUD
		case naluType == 39:
			t = types.H265_NALU_PREFIX_SEI
		case naluType == 42:
			t = types.H265_NALU_PREFIX_SEI
		case naluType == 62:
			t = types.H265_NALU_EOS
		case naluType == 37:
			t = types.H265_NALU_EOS
		default:
			t = types.NALU_UNKNOWN
		}

		info.NALUs = append(info.NALUs, t)

		if isKeyNALU {
			info.IsKey = true
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
