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
	maxLate         = 50
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
	sample := v.sb.Pop()
	return sample
}

func (v *VideoProcessor) GetFrameInfo(frame *media.Sample) *types.FrameInfo {
	if len(frame.Data) < 5 {
		return nil
	}

	frameInfo := &types.FrameInfo{}
	frameInfo.Codec = v.codec

	data := frame.Data
	// Пропускаем Annex-B Start Code (00 00 00 01 или 00 00 01)
	headerOffset := 0
	if data[0] == 0x00 && data[1] == 0x00 {
		if data[2] == 0x01 {
			headerOffset = 3
		} else if data[2] == 0x00 && data[3] == 0x01 {
			headerOffset = 4
		}
	}

	naluByte := data[headerOffset]

	if v.codec == types.H264 { // H264
		// Заголовок H264: [7 битов: 0] [1 бит: тип] -> берем 0x1F
		naluType := naluByte & 0x1F
		switch naluType {
		case 5:
			frameInfo.NALUType = types.H264_NALU_IDR
			frameInfo.IsKey = true
		case 1:
			frameInfo.NALUType = types.H264_NALU_NON_IDR
		case 7:
			frameInfo.NALUType = types.H264_NALU_SPS
		case 8:
			frameInfo.NALUType = types.H264_NALU_PPS
		case 6:
			frameInfo.NALUType = types.H264_NALU_SEI
		default:
			frameInfo.NALUType = types.NALU_UNKNOWN
		}
		frameInfo.NALUByte = naluType
	} else if v.codec == types.H265 {
		// H265
		// Заголовок H265: [1 бит: 0] [6 бит: тип] [3 бита: слой] [6 бит: ...]
		// Тип лежит в 1-м байте, сдвинутый на 1 бит вправо
		naluType := (naluByte >> 1) & 0x3F

		switch {
		case naluType == 1:
			frameInfo.NALUType = types.H265_NALU_TRAIL_R
		case naluType >= 16 && naluType <= 19:
			frameInfo.NALUType = types.H265_NALU_IDR_W
			frameInfo.IsKey = true
		case naluType == 20:
			frameInfo.NALUType = types.H265_NALU_IDR_N
			frameInfo.IsKey = true
		case naluType == 21:
			frameInfo.NALUType = types.H265_NALU_CRA
		case naluType == 32:
			frameInfo.NALUType = types.H265_NALU_VPS
		case naluType == 33:
			frameInfo.NALUType = types.H265_NALU_SPS
		case naluType == 34:
			frameInfo.NALUType = types.H265_NALU_PPS
		default:
			frameInfo.NALUType = types.NALU_UNKNOWN
		}
		frameInfo.NALUByte = naluType
	}
	return frameInfo
}
