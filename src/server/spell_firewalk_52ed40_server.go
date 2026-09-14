package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SpellFirewalkRuntime52ED40 supplies the flame effect still owned by the
// outer game runtime. The duration record and target never cross into PE32 C.
type SpellFirewalkRuntime52ED40 struct {
	SpawnFlame func(kind int, position types.Pointf)
}

// SpellFirewalkUpdate52ED40 binds the duration callback to native Go fields.
// Field72/76/80/84 retain the original four float32 bit patterns.
//
//go:noinline
func (s *Server) SpellFirewalkUpdate52ED40(record *DurSpell, runtime SpellFirewalkRuntime52ED40) int32 {
	return spellFirewalkUpdate52ED40(record, firewalkHooks52ED40[*DurSpell, *Object]{
		loadTarget:  func(record *DurSpell) *Object { return record.Target48 },
		loadFlags:   func(target *Object) uint32 { return uint32(target.ObjFlags) },
		loadFrame60: func(record *DurSpell) uint32 { return record.Frame60 },
		loadFrame64: func(record *DurSpell) uint32 { return record.Frame64 },
		storeFrame64: func(record *DurSpell, frame uint32) {
			record.Frame64 = frame
		},
		loadPosition: func(target *Object) types.Pointf { return target.PosVec },
		loadRadius:   func(target *Object) float32 { return target.Shape.Circle.R },
		loadPrevious: func(record *DurSpell) types.Pointf {
			return types.Pointf{
				X: math.Float32frombits(uint32(record.Field72)),
				Y: math.Float32frombits(uint32(record.Field76)),
			}
		},
		storePrevious: func(record *DurSpell, position types.Pointf) {
			record.Field72 = int32(math.Float32bits(position.X))
			record.Field76 = uintptr(math.Float32bits(position.Y))
		},
		loadAnchor: func(record *DurSpell) types.Pointf {
			return types.Pointf{
				X: math.Float32frombits(record.Field80),
				Y: math.Float32frombits(record.Field84),
			}
		},
		storeAnchor: func(record *DurSpell, position types.Pointf) {
			record.Field80 = math.Float32bits(position.X)
			record.Field84 = math.Float32bits(position.Y)
		},
		loadLevel:  func(record *DurSpell) uint32 { return record.Level },
		randomInt:  func(min, max int) int { return s.Rand.Logic.IntClamp(min, max) },
		spawnFlame: runtime.SpawnFlame,
	})
}
