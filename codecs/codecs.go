package codecs

import "time"

// NALUnitType — кастомный тип для удобства чтения
type NALUnitType uint8

const (
	// Примеры для H.265 (можно расширить)
	NAL_IDR_W_RADL NALUnitType = 19
	NAL_VPS        NALUnitType = 32
	NAL_SPS        NALUnitType = 33
	NAL_PPS        NALUnitType = 34
)

type CodecType string

const (
	H264 CodecType = "H264"
	H265 CodecType = "H265"
)

// VideoPacket представляет собой один NAL-юнит с метаданными
type VideoPacket struct {
	// Тип кодека (например, "H264" или "H265")
	Codec CodecType

	// Тип NAL-юнита (IDR, SPS, P-frame и т.д.)
	UnitType NALUnitType

	// Порядковый номер из RTP (тот самый seq=39270)
	SequenceNumber uint16

	// Временная метка из RTP (rtptime=39270)
	Timestamp uint32

	// Время получения пакета системой (для вычисления задержек)
	ArrivalTime time.Time

	// Сырые байты данных (без стартового кода 00 00 01)
	Payload []byte

	// Флаги для быстрой фильтрации
	IsKeyFrame bool // true если это IDR или CRA
	IsConfig   bool // true если это VPS/SPS/PPS
}

// MediaStream представляет собой поток данных
type MediaStream struct {
	TrackID int
	Packets chan VideoPacket
	// Здесь можно хранить актуальные параметры конфигурации
	LastSPS []byte
	LastPPS []byte
	LastVPS []byte // только для H265
}
