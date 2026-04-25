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

const (
	dialTimeout  = 5 * time.Second
	readDeadline = 3 * time.Second
)

type Client struct {
	conn        net.Conn
	reader      *bufio.Reader
	tp          *textproto.Reader
	csec        int
	digestAuth  auth.DigestAuth
	useUDP      bool
	rtpUDPConn  []*net.UDPConn
	rtcpUDPConn []*net.UDPConn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) RTPReader(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan RTSPResponse) error {
	defer func() {
		close(rtpCh)
	}()

	if c.useUDP {
		return c.readUDP(ctx, rtpCh, rtspCh)
	}

	// TCP interleaved
	go func() {
		<-ctx.Done()
		c.conn.Close()
	}()

	for {
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

		err = c.conn.SetReadDeadline(time.Now().Add(readDeadline))
		if err != nil {
			return err
		}
	}
}

func (c *Client) readUDP(ctx context.Context, rtpCh chan types.RTPPacket, rtspCh chan RTSPResponse) error {
	// Read RTSP from TCP
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				peek, err := c.reader.Peek(1)
				if err != nil {
					return
				}
				if peek[0] == 'R' {
					h, getHeadersErr := c.readRTSPHeaders()
					if getHeadersErr != nil {
						return
					}
					b, getBodyErr := c.readRTSPBody(h.Header)
					if getBodyErr != nil {
						return
					}
					rtspCh <- RTSPResponse{
						Headers: h,
						Body:    b,
					}
				}
			}
		}
	}()

	// Read RTP from UDP
	for i, conn := range c.rtpUDPConn {
		go func(i int, conn *net.UDPConn) {
			buf := make([]byte, 1500)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, _, err := conn.ReadFromUDP(buf)
					if err != nil {
						return
					}
					payload := make([]byte, n)
					copy(payload, buf[:n])
					rtpCh <- types.RTPPacket{
						Payload: payload,
						Channel: i * 2,
					}
				}
			}
		}(i, conn)
	}

	// Read RTCP from UDP
	for i, conn := range c.rtcpUDPConn {
		go func(i int, conn *net.UDPConn) {
			buf := make([]byte, 1500)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, _, err := conn.ReadFromUDP(buf)
					if err != nil {
						return
					}
					payload := make([]byte, n)
					copy(payload, buf[:n])
					rtpCh <- types.RTPPacket{
						Payload: payload,
						Channel: i*2 + 1,
					}
				}
			}
		}(i, conn)
	}

	<-ctx.Done()
	return nil
}

func (c *Client) Connect(ctx context.Context, u url.URL) error {
	c.csec = 1
	c.digestAuth = auth.DigestAuth{}

	d := net.Dialer{
		Timeout: dialTimeout,
	}

	conn, err := d.DialContext(ctx, "tcp", u.Host)
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
	if c.conn != nil {
		c.conn.Close()
	}
	for _, conn := range c.rtpUDPConn {
		if conn != nil {
			conn.Close()
		}
	}
	for _, conn := range c.rtcpUDPConn {
		if conn != nil {
			conn.Close()
		}
	}
	c.rtpUDPConn = nil
	c.rtcpUDPConn = nil
	return nil
}

func (c *Client) SetCredentials(credentials Credentials) {
	c.digestAuth.Username = credentials.Username
	c.digestAuth.Password = credentials.Password
}

func (c *Client) IsEmptyConnection() bool {
	return c.conn == nil
}

func (c *Client) SetUseUDP(use bool) {
	c.useUDP = use
}

func (c *Client) SetUDPConns(rtpConn, rtcpConn *net.UDPConn) {
	c.rtpUDPConn = append(c.rtpUDPConn, rtpConn)
	c.rtcpUDPConn = append(c.rtcpUDPConn, rtcpConn)
}
