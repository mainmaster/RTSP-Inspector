package rtsp_client

import (
	"rtsp-inspector/internal/types"
	"strings"

	"github.com/pixelbender/go-sdp/sdp"
)

type RTSPResponse struct {
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

func (res *RTSPResponse) GetTrackIDs() []string {
	trackIDs := make([]string, 0, 2)
	sdpSess, err := sdp.ParseString(string(res.Body))
	if err != nil {
		return trackIDs
	}
	for _, track := range sdpSess.Media {
		trackID := strings.Split(track.Attributes[0].Value, "=")[1]
		trackIDs = append(trackIDs, trackID)
	}
	return trackIDs
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

/*
func (res *RTSPResponse) GetChannelsMap() (map[types.RTPType]int, error) {
	channelsMap := make(map[types.RTPType]int)
	tracks := res.GetTrackIDs()
}
*/
