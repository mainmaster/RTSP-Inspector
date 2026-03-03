package processor

import (
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
		return nil
	}

	stats := &types.FrameInfo{}

	if codec == types.H264 { // H264
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
	} else if codec == types.H265 { // H265
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
