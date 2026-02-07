package rtsp_client

import (
	"fmt"
	"rtsp-inspector/internal/types"
	"strings"

	"github.com/pixelbender/go-sdp/sdp"
)

type RTSPResponse struct {
	Method types.RTSPMethod
	types.Headers
	Body []byte
}

func (res *RTSPResponse) GetSessionID() string {
	session := res.Header.Get("Session")
	if session != "" {
		session = strings.Split(session, ";")[0]
	}
	return session
}

func (res *RTSPResponse) GetTrackIDs() ([]string, error) {
	trackIDs := make([]string, 0, 2)

	sdpSess, err := sdp.ParseString(string(res.Body))
	if err != nil {
		return nil, err
	}
	for _, track := range sdpSess.Media {
		for _, attr := range track.Attributes {
			if attr.Name == "control" {
				trackID := strings.Split(attr.Value, "=")[1]
				trackIDs = append(trackIDs, trackID)
			}
		}
	}
	if len(trackIDs) == 0 {
		return trackIDs, fmt.Errorf("no track IDs found in RTSP response")
	}
	return trackIDs, nil
}

func (res *RTSPResponse) GetCodecs() (map[string]types.CodecType, error) {
	codecMap := make(map[string]types.CodecType)
	sdpSess, err := sdp.ParseString(string(res.Body))
	if err != nil {
		return nil, err
	}
	for _, m := range sdpSess.Media {
		switch strings.ToLower(m.Format[0].Name) {
		case "h264":
			codecMap[m.Type] = types.H264
		case "h265":
			codecMap[m.Type] = types.H265
		}
	}
	return codecMap, nil
}
