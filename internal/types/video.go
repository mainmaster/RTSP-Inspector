package types

type CodecType int

const (
	H264 CodecType = iota
	H265
)

type NALUType int

const (
	NALU_UNKNOWN NALUType = iota

	H264_NALU_NON_IDR
	H264_NALU_IDR
	H264_NALU_SPS
	H264_NALU_PPS
	H264_NALU_SEI

	H265_NALU_TSA_N      // Тип 2 (Temporal Sub-layer Access)
	H265_NALU_TSA_R      // Тип 3
	H265_NALU_STSA_N     // Тип 6
	H265_NALU_STSA_R     // Тип 7      // TSA Temporal Sub-layer Access
	H265_NALU_TRAIL_R    // Обычный P/B
	H265_NALU_TRAIL_N    // Тип 0 (Non-Ref)
	H265_NALU_RASL_R     // Тип 4 (Leading)
	H265_NALU_RADL_R     // Тип 5
	H265_NALU_IDR_W      // Ключ (BLA_W_LP / IDR_W_RADL)
	H265_NALU_IDR_N      // Ключ (IDR_N_LP)
	H265_NALU_CRA        // Точка входа
	H265_NALU_IDR_N_LP   // Тип 23 (Твой случай)
	H265_NALU_IDR_W_RADL // Тип 25

	H265_NALU_VPS
	H265_NALU_SPS
	H265_NALU_PPS
	H265_NALU_AUD        // Тип 35 (Разделитель)
	H265_NALU_EOS        // Тип 36/37/62 (Конец потока)
	H265_NALU_PREFIX_SEI // Тип 39
	H265_NALU_SUFFIX_SEI // Тип 40 / 37

	H265_NALU_RSV_NVCL42 // Тип 42
	H265_NALU_RSV_NVCL43 // Тип 43
	H265_NALU_SKIP       // Для типов 6, 9, 10 и т.д.
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
	H265_NALU_TSA_N:      "TSA_N (Temporal Access)",
	H265_NALU_TSA_R:      "TSA_R (Temporal Access)",
	H265_NALU_STSA_N:     "STSA_N (Step-wise Temporal)",
	H265_NALU_STSA_R:     "STSA_R (Step-wise Temporal)",
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

type NALUInfo struct {
	Type  NALUType
	IsKey bool
}

var H265TypeLookup = map[int]NALUInfo{
	0:  {H265_NALU_TRAIL_N, false},
	1:  {H265_NALU_TRAIL_R, false},
	2:  {H265_NALU_TSA_N, false},
	3:  {H265_NALU_TSA_R, false},
	4:  {H265_NALU_RASL_R, false},
	5:  {H265_NALU_RADL_R, false},
	6:  {H265_NALU_STSA_N, false},
	7:  {H265_NALU_STSA_R, false},
	9:  {H265_NALU_SKIP, false},
	16: {H265_NALU_IDR_W, true},
	17: {H265_NALU_IDR_W, true},
	18: {H265_NALU_IDR_W, true},
	19: {H265_NALU_IDR_W, true},
	20: {H265_NALU_IDR_N, true},
	21: {H265_NALU_CRA, true},
	22: {H265_NALU_IDR_N_LP, true},
	23: {H265_NALU_IDR_N_LP, true},
	24: {H265_NALU_IDR_N_LP, true},
	25: {H265_NALU_IDR_W_RADL, true},

	// --- Non-VCL NAL Units (Служебные) ---
	32: {H265_NALU_VPS, false},        // VPS
	33: {H265_NALU_SPS, false},        // SPS
	34: {H265_NALU_PPS, false},        // PPS
	35: {H265_NALU_AUD, false},        // AUD (Разделитель)
	36: {H265_NALU_EOS, false},        // End of Sequence
	37: {H265_NALU_EOS, false},        // End of Bitstream (Тип 37/62)
	39: {H265_NALU_PREFIX_SEI, false}, // Prefix SEI
	40: {H265_NALU_SUFFIX_SEI, false}, // Suffix SEI
	42: {H265_NALU_RSV_NVCL42, false}, // Тот самый тип 42
	43: {H265_NALU_RSV_NVCL43, false}, // Тот самый тип 43
	62: {H265_NALU_EOS, false},        // Еще один EOS
}

var H264TypeLookup = map[int]NALUInfo{
	1: {H264_NALU_NON_IDR, false},
	5: {H264_NALU_IDR, true},
	6: {H264_NALU_SEI, false},
	7: {H264_NALU_SPS, false},
	8: {H264_NALU_PPS, false},
}

type FrameInfo struct {
	Codec CodecType
	NALUs []NALUType
	IsKey bool
}
