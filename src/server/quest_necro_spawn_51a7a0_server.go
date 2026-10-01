package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// QuestNecroSpawnRuntime51A7A0 binds the original scalar global and the outer
// server's creation/inventory services. All object/update/health/position
// references stay native pointers, never PE32 words.
type QuestNecroSpawnRuntime51A7A0 struct {
	HealthScale  func() float64
	SetHP        func(*Object, uint16)
	CreateAt     func(*Object, types.Pointf)
	Stage        func() uint32
	InventoryPut func(*Object, *Object, int32)
}

func (s *Server) questNecroSpawnNativeHooks51A7A0(position *types.Pointf, r QuestNecroSpawnRuntime51A7A0) questNecroSpawnHooks51A7A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf] {
	return questNecroSpawnHooks51A7A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf]{
		newObject:       s.NewObjectByTypeID,
		healthScale:     r.HealthScale,
		loadUpdate:      func(unit *Object) *MonsterUpdateData { return (*MonsterUpdateData)(unit.UpdateData) },
		loadDefinition:  func(update *MonsterUpdateData) *MonsterDef { return update.MonsterDef },
		loadQuestHealth: func(definition *MonsterDef) int32 { return int32(definition.HealthQuest72) },
		loadType:        func(unit *Object) uint16 { return unit.TypeInd },
		lookupType:      func(id uint16) *ObjectType { return s.Types.ByInd(int(id)) },
		loadTypeHealth:  (*ObjectType).Health,
		loadHealth:      func(unit *Object) *HealthData { return unit.HealthData },
		loadCurrent:     func(health *HealthData) uint16 { return health.Cur },
		loadMaximum:     func(health *HealthData) uint16 { return health.Max },
		setHP:           r.SetHP,
		storeMaximum:    func(health *HealthData, value uint16) { health.Max = value },
		storeAI: func(update *MonsterUpdateData, field, value uint32) {
			switch field {
			case 340:
				update.AIAction340 = value
			case 411:
				update.Field411 = value
			case 423:
				update.Field423 = value
			case 326:
				update.Aggression = math.Float32frombits(value)
			case 510:
				update.Field510 = value
			case 410:
				update.Field410 = value
			case 444:
				update.Field444 = value
			case 415:
				update.Field415 = value
			default:
				panic("invalid Quest necromancer AI field")
			}
		},
		loadPosition:   func() types.Pointf { return *position },
		createAt:       r.CreateAt,
		stage:          r.Stage,
		activateReward: s.RewardMarkerActivateDefault4F0720,
		inventoryPut:   r.InventoryPut,
		freeObject:     func(obj *Object) { s.Objs.FreeObject(obj) },
	}
}

// QuestSpawnNecro51A7A0 creates the stock Quest minion and one reward.
func (s *Server) QuestSpawnNecro51A7A0(position *types.Pointf, r QuestNecroSpawnRuntime51A7A0) {
	questNecroSpawn51A7A0(s.questNecroSpawnNativeHooks51A7A0(position, r))
}
