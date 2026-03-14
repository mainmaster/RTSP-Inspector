package rtsp_client

import (
	"encoding/binary"

	"github.com/pion/rtcp"
)

func (c *Client) SendPLI(rtcpChannel byte, ssrc uint32) error {
	pli := &rtcp.PictureLossIndication{
		MediaSSRC: ssrc,
	}

	data, err := pli.Marshal()
	if err != nil {
		return err
	}

	// Формируем заголовок RTSP Interleaved ($ + channel + 2 bytes length)
	header := make([]byte, 4)
	header[0] = '$'
	header[1] = rtcpChannel
	binary.BigEndian.PutUint16(header[2:], uint16(len(data)))

	_, err = c.conn.Write(append(header, data...))
	return err
}
