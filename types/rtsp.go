package types

type DataChannels struct {
	VideoCh chan []byte // Видео (RTP)
	AudioCh chan []byte // Звук (RTP) - опционально
	RTCPCh  chan []byte // Статистика/Контроль
	ErrCh   chan error  // Ошибки сети и парсинга
}
