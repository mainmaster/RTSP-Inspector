package types

type PacketCounter struct {
	All       int
	Video     int
	Audio     int
	RTCPVideo int
	RTCPAudio int
}
