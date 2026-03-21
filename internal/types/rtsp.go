package types

import "net/textproto"

type TrackType int

const (
	TrackTypeAudio TrackType = iota
	TrackTypeVideo
)

type Track struct {
	TrackType TrackType
	ID        string
}

type Interleaved struct {
	TrackType           TrackType
	InterleavedChannels []int
}

type RTSPMethod string

const (
	MethodOptions  RTSPMethod = "OPTIONS"
	MethodDescribe RTSPMethod = "DESCRIBE"
	MethodSetup    RTSPMethod = "SETUP"
	MethodPlay     RTSPMethod = "PLAY"
	MethodPause    RTSPMethod = "PAUSE"
	MethodTeardown RTSPMethod = "TEARDOWN"
)

type Headers struct {
	StatusCode int
	StatusLine string
	Header     textproto.MIMEHeader
}

type Header map[string]string

func (h Header) Add(key, value string) {
	h[key] = value
}
