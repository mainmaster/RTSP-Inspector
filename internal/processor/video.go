package processor

import (
	"fmt"
	"rtsp-inspector/internal/types"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

func getDepacketizer(codecType types.CodecType) rtp.Depacketizer {
	var depacketizer rtp.Depacketizer
	switch codecType {
	case types.H264:
		depacketizer = &codecs.H264Packet{}
	case types.H265:
		depacketizer = &codecs.H265Depacketizer{}
	}
	return depacketizer
}

func GetFrameInfo(payload []byte, codec types.CodecType) *types.FrameInfo {
	depacketizer := getDepacketizer(codec)
	nalu, err := depacketizer.Unmarshal(payload)
	if err != nil || len(nalu) == 0 {
		fmt.Println(err)
		return nil
	}

	frameInfo := &types.FrameInfo{}
	frameInfo.Codec = codec

	if codec == types.H264 { // H264
		// Заголовок H264: [7 битов: 0] [1 бит: тип] -> берем 0x1F
		naluType := nalu[0] & 0x1F
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
	} else if codec == types.H265 {
		if len(nalu) < 1 {
			return nil
		}
		// H265
		// Заголовок H265: [1 бит: 0] [6 бит: тип] [3 бита: слой] [6 бит: ...]
		// Тип лежит в 1-м байте, сдвинутый на 1 бит вправо
		naluType := (nalu[0] >> 1) & 0x3F
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
	}
	return frameInfo
}
