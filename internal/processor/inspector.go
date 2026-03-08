package processor

import (
	"rtsp-inspector/internal/types"
	"sync"
)

type StreamInspector struct {
	mu            sync.Mutex
	packetCounter map[types.RTPType]int
	naluCounter   map[types.NALUType]int
}

func NewStreamInspector() *StreamInspector {
	return &StreamInspector{
		packetCounter: make(map[types.RTPType]int),
		naluCounter:   make(map[types.NALUType]int),
		mu:            sync.Mutex{},
	}
}

func (si *StreamInspector) IncrementNALUCounter(nalus []types.NALUType) {
	si.mu.Lock()
	defer si.mu.Unlock()

	for _, n := range nalus {
		si.naluCounter[n]++
	}
}

func (si *StreamInspector) IncrementRTPCounter(packet *types.RTPPacket) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.packetCounter[packet.Type]++
}

func (si *StreamInspector) GetRTPCounter() map[types.RTPType]int {
	si.mu.Lock()
	defer si.mu.Unlock()
	newMap := make(map[types.RTPType]int)
	for k, v := range si.packetCounter {
		newMap[k] = v
	}
	return newMap
}

func (si *StreamInspector) GetNALUCounter() map[types.NALUType]int {
	si.mu.Lock()
	defer si.mu.Unlock()
	newMap := make(map[types.NALUType]int)
	for k, v := range si.naluCounter {
		newMap[k] = v
	}
	return newMap
}

func (si *StreamInspector) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	clear(si.packetCounter)
	clear(si.naluCounter)
}
