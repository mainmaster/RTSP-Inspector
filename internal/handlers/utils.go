package handlers

import (
	"fmt"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/internal/types"
	"strings"
)

func getMapFromInterleaved(rtspRes *RTSPFlowResponse) map[int]types.RTPType {
	interleavedMap := make(map[int]types.RTPType)
	for _, i := range rtspRes.Interleaved {
		switch i.TrackType {
		case types.TrackTypeVideo:
			interleavedMap[i.InterleavedChannels[0]] = types.RTPTypeVideo
			interleavedMap[i.InterleavedChannels[1]] = types.RTCPTypeVideo

		case types.TrackTypeAudio:
			interleavedMap[i.InterleavedChannels[0]] = types.RTPTypeAudio
			interleavedMap[i.InterleavedChannels[1]] = types.RTCPTypeAudio
		default:
		}
	}
	return interleavedMap
}

func buildOutputString(response *rtsp_client.RTSPResponse) string {
	var output strings.Builder
	output.WriteString(response.StatusLine)
	output.WriteString("\r\n")
	for k, v := range response.Header {
		output.WriteString(fmt.Sprintf("%s: %s", k, v[0]))
		output.WriteString("\r\n")
	}
	output.WriteString("\r\n")
	output.WriteString(string(response.Body))
	return output.String()
}
