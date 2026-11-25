package server

import "encoding/binary"

const (
	NetworkVoteControlPacketSize51BAD0 = 2
	NetworkVoteNamedPacketSize51BAD0   = 52
)

// DecodeVotePacket51BAD0 returns the name with a spare terminator for the
// legacy lookup. A zero length signals a missing or unrecognized vote.
func DecodeVotePacket51BAD0(data []byte, quest bool) (kind int, cancel bool, name []uint16, size int) {
	if len(data) < NetworkVoteControlPacketSize51BAD0 {
		return 0, false, nil, 0
	}
	subtype := data[1]
	switch subtype {
	case 0, 1, 2, 3:
		if len(data) < NetworkVoteNamedPacketSize51BAD0 {
			return 0, false, nil, 0
		}
		const units = (NetworkVoteNamedPacketSize51BAD0 - NetworkVoteControlPacketSize51BAD0) / 2
		name = make([]uint16, units+1)
		for i := range units {
			name[i] = binary.LittleEndian.Uint16(data[2+2*i : 4+2*i])
		}
		if subtype&1 != 0 {
			kind = 1
		} else if quest {
			kind = 3
		}
		return kind, subtype >= 2, name, NetworkVoteNamedPacketSize51BAD0
	case 4, 5:
		return 2, subtype == 5, nil, NetworkVoteControlPacketSize51BAD0
	default:
		return 0, false, nil, 0
	}
}
