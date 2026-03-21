package rtsp_client

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"rtsp-inspector/internal/rtsp_client/auth"
	"rtsp-inspector/internal/types"
	"time"
)

type Client struct {
	conn       net.Conn
	reader     *bufio.Reader
	tp         *textproto.Reader
	csec       int
	digestAuth auth.DigestAuth
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) RTPReader(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan RTSPResponse) error {
	defer func() {
		close(rtpCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		peek, err := c.reader.Peek(1)
		if err != nil {
			return err // EOF
		}

		switch peek[0] {
		case '$': // RTP
			_, err = c.reader.Discard(1)
			if err != nil {
				return err
			}
			channelByte, _ := c.reader.ReadByte()

			lenBuf := make([]byte, 2)
			if _, err = io.ReadFull(c.reader, lenBuf); err != nil {
				return err
			}
			length := binary.BigEndian.Uint16(lenBuf)
			payload := make([]byte, length)
			if _, err = io.ReadFull(c.reader, payload); err != nil {
				return err
			}

			rtpCh <- types.RTPPacket{
				Payload: payload,
				Channel: int(channelByte),
			}
		case 'R': // RTSP
			h, getHeadersErr := c.readRTSPHeaders()
			if getHeadersErr != nil {
				return getHeadersErr
			}
			b, getBodyErr := c.readRTSPBody(h.Header)
			if getBodyErr != nil {
				return getBodyErr
			}

			rtspCh <- RTSPResponse{
				Headers: h,
				Body:    b,
			}

		default:
			return errors.New("RTP header not supported")
		}

		err = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err != nil {
			return err
		}
	}
}

func (c *Client) Connect(u url.URL) error {
	c.csec = 1
	c.digestAuth = auth.DigestAuth{}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.tp = textproto.NewReader(c.reader)

	c.setCredentialsFromURL(u)

	if c.digestAuth.Password != "" && c.digestAuth.Nonce == "" && c.digestAuth.Realm == "" {
		return c.setDigestHeaders(u)
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SetCredentials(credentials Credentials) {
	c.digestAuth.Username = credentials.Username
	c.digestAuth.Password = credentials.Password
}

func (c *Client) IsEmptyConnection() bool {
	if c.conn == nil {
		return true
	}
	return false
}
