package types

import (
	"encoding/binary"
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
	Channel int
}

var RTPTypeNames = map[RTPType]string{
	RTPTypeVideo:  "Video",
	RTCPTypeVideo: "RTCP Video",
	RTPTypeAudio:  "Audio",
	RTCPTypeAudio: "RTCP Audio",
}

func (t RTPPacket) GetSSRC() uint32 {
	videoSSRC := binary.BigEndian.Uint32(t.Payload[8:12])
	return videoSSRC
}
