package types

import (
	"net/textproto"
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

type DataChannels struct {
	VideoCh chan []byte // Видео (RTP)
	AudioCh chan []byte // Звук (RTP) - опционально
	RTCPCh  chan []byte // Статистика/Контроль
	ErrCh   chan error  // Ошибки сети и парсинга
}
