package rtsp_client

import (
	"net/textproto"
	"rtsp-inspector/internal/types"
	"strings"

	"github.com/pixelbender/go-sdp/sdp"
)

type Headers struct {
	StatusCode int
	StatusLine string
	Header     textproto.MIMEHeader
}

type Response struct {
	Headers
	Body []byte
}

func (res *Response) GetSessionID() string {
	session := res.Header.Get("Session")
	if session != "" {
		session = strings.Split(session, ";")[0]
	}
	return session
}

func (res *Response) GetTrackIDs() []string {
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

func (res *Response) GetCodecs() (map[string]types.CodecType, error) {
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
