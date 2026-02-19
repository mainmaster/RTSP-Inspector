package rtsp

import (
	"net/textproto"
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
