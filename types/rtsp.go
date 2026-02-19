package types

type DataChannels struct {
	VideoCh     chan []byte
	AudioCh     chan []byte
	RTCPVideoCh chan []byte
	RTCPAudioCh chan []byte
}
