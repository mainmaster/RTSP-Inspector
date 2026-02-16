package codecs

/*

package main

import (
	"fmt"
	"time"
)

// Константы для кодеков
const (
	CodecH264 = "H264"
	CodecH265 = "H265"
)

// ParseNALUnit анализирует заголовок и создает структуру VideoPacket
func ParseNALUnit(payload []byte, codec string, seq uint16, rtptime uint32) VideoPacket {
	packet := VideoPacket{
		Codec:          codec,
		SequenceNumber: seq,
		Timestamp:      rtptime,
		ArrivalTime:    time.Now(),
		Payload:        payload,
	}

	if len(payload) == 0 {
		return packet
	}

	if codec == CodecH264 {
		// H.264: тип NAL содержится в первом байте (последние 5 бит)
		// Формат: F (1 бит) | NRI (2 бита) | Type (5 бит)
		packet.UnitType = NALUnitType(payload[0] & 0x1F)

		// 7 = SPS, 8 = PPS, 5 = IDR (Ключевой кадр)
		packet.IsConfig = (packet.UnitType == 7 || packet.UnitType == 8)
		packet.IsKeyFrame = (packet.UnitType == 5)

	} else if codec == CodecH265 {
		// H.265: тип NAL содержится в первом и втором байтах
		// Формат: F (1 бит) | Type (6 бит) | LayerID (6 бит) | TID (3 бита)
		// Сдвигаем первый байт на 1 вправо и убираем верхний бит (0x3F = 00111111)
		packet.UnitType = NALUnitType((payload[0] >> 1) & 0x3F)

		// 32=VPS, 33=SPS, 34=PPS
		packet.IsConfig = (packet.UnitType >= 32 && packet.UnitType <= 34)
		// 19=IDR_W_RADL, 20=IDR_N_LP, 21=CRA_NUT
		packet.IsKeyFrame = (packet.UnitType >= 19 && packet.UnitType <= 21)
	}

	return packet
}

func main() {
	// Пример: заголовок H.265 SPS пакета (0x42 0x01...)
	// 0x42 в двоичном виде: 01000010
	// Сдвиг >> 1: 00100001 (это число 33 в десятичной системе)
	rawH265SPS := []byte{0x42, 0x01, 0x01}

	pkt := ParseNALUnit(rawH265SPS, CodecH265, 39270, 39270)

	fmt.Printf("Кодек: %s\n", pkt.Codec)
	fmt.Printf("Тип NAL: %d\n", pkt.UnitType)
	fmt.Printf("Конфиг: %v\n", pkt.IsConfig)
	fmt.Printf("Ключевой кадр: %v\n", pkt.IsKeyFrame)
}

*/
