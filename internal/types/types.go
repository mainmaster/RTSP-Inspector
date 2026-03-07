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

	H265_NALU_TRAIL_R    // P/B
	H265_NALU_IDR_W      // KEY
	H265_NALU_IDR_N      // KEY (no leading)
	H265_NALU_CRA        // CRA
	H265_NALU_VPS        // VPS
	H265_NALU_SPS        // SPS
	H265_NALU_PPS        // PPS
	H265_NALU_PREFIX_SEI // Prefix SEI
)

var NALUNames = map[NALUType]string{
	NALU_UNKNOWN:         "UNKNOWN",
	H264_NALU_NON_IDR:    "Non-IDR (P/B-Frame)",
	H264_NALU_IDR:        "IDR",
	H264_NALU_SPS:        "SPS",
	H264_NALU_PPS:        "PPS",
	H264_NALU_SEI:        "SEI",
	H265_NALU_TRAIL_R:    "TRAIL_R (P/B-Frame)",
	H265_NALU_IDR_W:      "IDR_W",
	H265_NALU_IDR_N:      "IDR_N",
	H265_NALU_CRA:        "CRA",
	H265_NALU_VPS:        "VPS",
	H265_NALU_SPS:        "SPS",
	H265_NALU_PPS:        "PPS",
	H265_NALU_PREFIX_SEI: "PREFIX_SEI",
}

type FrameInfo struct {
	Codec CodecType
	NALUs []NALUType
	IsKey bool
}
