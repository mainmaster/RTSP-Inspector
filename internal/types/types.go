package types

import "github.com/pion/rtp"

type RTPType int

const (
	RTPTypeVideo RTPType = iota
	RTCPTypeVideo
	RTPTypeAudio
	RTCPTypeAudio
)

type RTPPacket struct {
	Packet *rtp.Packet
	Type   RTPType
}

type CodecType int

const (
	H264 CodecType = iota
	H265
)

type NALUType int

const (
	H264_NALU_NON_IDR NALUType = 1
	H264_NALU_IDR     NALUType = 5
	H264_NALU_SPS     NALUType = 7
	H264_NALU_PPS     NALUType = 8
	H264_NALU_SEI     NALUType = 6

	H265_NALU_TRAIL_R NALUType = 1  // Обычный P/B кадр
	H265_NALU_IDR_W   NALUType = 19 // Ключевой кадр
	H265_NALU_IDR_N   NALUType = 20 // Ключевой кадр (no leading)
	H265_NALU_CRA     NALUType = 21 // Точка входа (CRA)
	H265_NALU_VPS     NALUType = 32
	H265_NALU_SPS     NALUType = 33
	H265_NALU_PPS     NALUType = 34

	NALU_UNKNOWN NALUType = 0
)

type FrameInfo struct {
	Codec   string
	Type    string
	IsKey   bool
	NaluRaw int
}
