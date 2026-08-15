package legacy

import "github.com/opennox/opennox/v1/server"

// respawnRecord4EC5E0 is the pointer-width-normalized Go view of
// nox_respawn_record_t. Its native offsets are sealed by the architecture
// specific compile-time assertions next to this file and by defs.h.
type respawnRecord4EC5E0 struct {
	TypeInd    uint32
	Object     *server.Object
	X          float32
	Y          float32
	Direction  uint16
	Reserved18 uint16
	RespawnAt  uint32
	Pending    uint32
	Attrs      server.ModifierInitData
	Charge1    uint8
	Charge0    uint8
	Reserved50 uint16
	Next       *respawnRecord4EC5E0
	Prev       *respawnRecord4EC5E0
}
