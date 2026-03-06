package types

type RTPType int

const (
	RTPTypeVideo RTPType = iota
	RTCPTypeVideo
	RTPTypeAudio
	RTCPTypeAudio
)

type RTPPacket struct {
	Payload []byte
	Type    RTPType
}

type CodecType int

const (
	H264 CodecType = iota
	H265
)

type NALUType int

const (
	NALU_UNKNOWN      NALUType = iota
	H264_NALU_NON_IDR          // Non-IDR (P/B-Frame)
	H264_NALU_IDR              // IDR (I-Frame)
	H264_NALU_SPS              // SPS
	H264_NALU_PPS              // PPS
	H264_NALU_SEI              // SEI

	H265_NALU_TRAIL_R // Обычный P/B кадр
	H265_NALU_IDR_W   // Ключевой кадр
	H265_NALU_IDR_N   // Ключевой кадр (no leading)
	H265_NALU_CRA     // Точка входа (CRA)
	H265_NALU_VPS     // VPS
	H265_NALU_SPS     // SPS
	H265_NALU_PPS     // PPS
)

type FrameInfo struct {
	Codec    CodecType
	NALUType NALUType
	NALUByte byte
	IsKey    bool
}
