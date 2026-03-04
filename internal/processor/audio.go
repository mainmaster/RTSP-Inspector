package processor

import (
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

type AudioProcessor struct {
	client       *rtsp_client.Client
	codecType    types.CodecType
	depacketizer rtp.Depacketizer
}
type AudioDataChannels struct {
	Ch     chan []byte
	RTCPCh chan []rtcp.Packet
}
