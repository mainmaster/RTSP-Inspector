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

var RTPTypeNames = map[RTPType]string{
	RTPTypeVideo:  "Video",
	RTCPTypeVideo: "RTCP Video",
	RTPTypeAudio:  "Audio",
	RTCPTypeAudio: "RTCP Audio",
}

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
	H265_NALU_EOS        // End of stream
	H265_NALU_SUFFIX_SEI //Suffix SEI
	H265_NALU_RASL_R     // RASL_R (Leading Frame)
	H265_NALU_RSV_NVCL42 // Reserved Non-VCL (Type 42)
	H265_NALU_IDR_N_LP   // IDR_N_LP (Pure Keyframe)
	H265_NALU_IDR_W_RADL // IDR_W_RADL (Keyframe + Leading)
	H265_NALU_AUD        // AUD (Access Unit Delimiter)",
)

var NALUNames = map[NALUType]string{
	NALU_UNKNOWN: "UNKNOWN",

	// H.264 (AVC)
	H265_NALU_IDR_N_LP: "IDR_N_LP (Pure Keyframe)",
	H264_NALU_NON_IDR:  "P/B-Frame (Delta)",
	H264_NALU_IDR:      "IDR Frame (Keyframe)",
	H264_NALU_SPS:      "SPS (Sequence)",
	H264_NALU_PPS:      "PPS (Picture)",
	H264_NALU_SEI:      "SEI Metadata",

	// H.265 (HEVC)
	H265_NALU_TRAIL_R:    "P/B-Frame (Trailing)",
	H265_NALU_IDR_W:      "IDR_W (Keyframe)",
	H265_NALU_IDR_N:      "IDR_N (Keyframe)",
	H265_NALU_CRA:        "CRA (Clean Access)",
	H265_NALU_VPS:        "VPS (Video)",
	H265_NALU_SPS:        "SPS (Sequence)",
	H265_NALU_PPS:        "PPS (Picture)",
	H265_NALU_PREFIX_SEI: "Prefix SEI Metadata",
	H265_NALU_SUFFIX_SEI: "Suffix SEI (Post-Metadata)",
	H265_NALU_EOS:        "EOS (End of stream)",
	H265_NALU_RASL_R:     "RASL_R (Leading Frame)",
	H265_NALU_RSV_NVCL42: "Reserved Non-VCL (Type 42)",
	H265_NALU_IDR_W_RADL: "IDR_W_RADL (Keyframe + Leading)",
	H265_NALU_AUD:        "AUD (Access Unit Delimiter)",
}

type FrameInfo struct {
	Codec CodecType
	NALUs []NALUType
	IsKey bool
}
