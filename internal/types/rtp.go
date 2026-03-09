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

var RTPTypeNames = map[RTPType]string{
	RTPTypeVideo:  "Video",
	RTCPTypeVideo: "RTCP Video",
	RTPTypeAudio:  "Audio",
	RTCPTypeAudio: "RTCP Audio",
}
