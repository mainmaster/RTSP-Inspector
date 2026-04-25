package handlers

import "rtsp-inspector/internal/types"

type UIUpdater interface {
	AddLogEntry(title types.RTSPMethod, body string, isRequest bool)
	ShowError(err error)
	UpdateConnectStatus(isConnected bool)
	GetURL() string
	GetTransport() string
	UpdateRTPCounter(counter map[types.RTPType]int)
	UpdateNALUCounter(counter map[types.NALUType]int)
	ClearCounters()
	ClearLogs()
}
