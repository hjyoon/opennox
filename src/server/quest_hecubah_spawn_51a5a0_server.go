package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// Both stock Quest minion creators use the same native server services.
type QuestHecubahSpawnRuntime51A5A0 = QuestNecroSpawnRuntime51A7A0

func (s *Server) questHecubahSpawnNativeHooks51A5A0(position *types.Pointf, r QuestHecubahSpawnRuntime51A5A0) questHecubahSpawnHooks51A5A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf] {
	h := questHecubahSpawnHooks51A5A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf]{
		questNecroSpawnHooks51A7A0: s.questNecroSpawnNativeHooks51A7A0(position, r),
		balanceFloat:               func(key string) float32 { return float32(s.Balance.Float(key)) },
	}
	storeAI := h.storeAI
	h.storeAI = func(update *MonsterUpdateData, field, value uint32) {
		switch field {
		case 388:
			update.Field388 = value
		case 330:
			update.Field330 = math.Float32frombits(value)
		default:
			storeAI(update, field, value)
		}
	}
	return h
}

// QuestSpawnHecubah51A5A0 creates the stock Quest boss and four rewards.
func (s *Server) QuestSpawnHecubah51A5A0(position *types.Pointf, r QuestHecubahSpawnRuntime51A5A0) {
	questHecubahSpawn51A5A0(s.questHecubahSpawnNativeHooks51A5A0(position, r))
}
