package types

type RTSPMethod string

const (
	MethodOptions  RTSPMethod = "OPTIONS"
	MethodDescribe RTSPMethod = "DESCRIBE"
	MethodSetup    RTSPMethod = "SETUP"
	MethodPlay     RTSPMethod = "PLAY"
	MethodPause    RTSPMethod = "PAUSE"
	MethodTeardown RTSPMethod = "TEARDOWN"
)

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

var NALUNames = map[NALUType]string{
	NALU_UNKNOWN:      "UNKNOWN",
	H264_NALU_NON_IDR: "H264_NON_IDR",
	H264_NALU_IDR:     "H264_IDR",
	H264_NALU_SPS:     "H264_SPS",
	H264_NALU_PPS:     "H264_PPS",
	H264_NALU_SEI:     "H264_SEI",
	H265_NALU_TRAIL_R: "H265_TRAIL_R",
	H265_NALU_IDR_W:   "H265_IDR_W",
	H265_NALU_IDR_N:   "H265_IDR_N",
	H265_NALU_CRA:     "H265_CRA",
	H265_NALU_VPS:     "H265_VPS",
	H265_NALU_SPS:     "H265_SPS",
	H265_NALU_PPS:     "H265_PPS",
}

type FrameInfo struct {
	Codec CodecType
	NALUs []NALUType
	IsKey bool
}
