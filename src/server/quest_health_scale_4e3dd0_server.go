package server

import "unsafe"

// QuestHealthScaleRuntime4E3DD0 contains scalar PE32 globals and callbacks to
// retained scalar/set-HP routines. All object/type/update/health links stay
// native-width typed pointers, including links reloaded after set-HP.
type QuestHealthScaleRuntime4E3DD0 struct {
	Ready                                            *uint32
	DamageInit, HealthInit, DamageCoeff, HealthCoeff *float32
	Difficulty                                       func() float64
	StoreDamageScale, StoreHealthScale               func(float32)
	HealthScale                                      func() float64
	SetHP                                            func(*Object, uint16) int32
}

func (s *Server) questHealthScaleNativeHooks4E3DD0(r QuestHealthScaleRuntime4E3DD0) questHealthScaleHooks4E3DD0[*Object, *MonsterDef, *HealthData, *ObjectType, *MonsterUpdateData] {
	cache := [4]*float32{r.DamageInit, r.HealthInit, r.DamageCoeff, r.HealthCoeff}
	return questHealthScaleHooks4E3DD0[*Object, *MonsterDef, *HealthData, *ObjectType, *MonsterUpdateData]{
		balanceFloat: func(key string) float32 { return float32(s.Balance.Float(key)) },
		loadReady:    func() uint32 { return *r.Ready },
		storeReady:   func(value uint32) { *r.Ready = value },
		loadCache:    func(index questHealthCache4E3DD0) float32 { return *cache[index] },
		storeCache:   func(index questHealthCache4E3DD0, value float32) { *cache[index] = value },
		difficulty:   r.Difficulty, storeDamageScale: r.StoreDamageScale,
		storeHealthScale: r.StoreHealthScale, healthScale: r.HealthScale,
		first: s.Objs.First, next: (*Object).Next,
		loadClass:  func(obj *Object) uint32 { return uint32(obj.ObjClass) },
		loadFlags:  func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		loadHealth: func(obj *Object) *HealthData { return obj.HealthData },
		healthPointerWord: func(health *HealthData) uint16 {
			return uint16(uintptr(unsafe.Pointer(health)))
		},
		loadCurrent:     func(health *HealthData) uint16 { return health.Cur },
		loadMaximum:     func(health *HealthData) uint16 { return health.Max },
		loadTypeID:      func(obj *Object) uint16 { return obj.TypeInd },
		lookupType:      func(id uint16) *ObjectType { return s.Types.ByInd(int(id)) },
		loadTypeHealth:  (*ObjectType).Health,
		loadUpdate:      func(obj *Object) *MonsterUpdateData { return (*MonsterUpdateData)(obj.UpdateData) },
		loadDefinition:  func(update *MonsterUpdateData) *MonsterDef { return update.MonsterDef },
		loadQuestHealth: func(definition *MonsterDef) uint16 { return uint16(definition.HealthQuest72) },
		loadStatusByte:  func(update *MonsterUpdateData) uint8 { return uint8(update.StatusFlags) },
		setHP:           r.SetHP,
		storeMaximum:    func(health *HealthData, value uint16) { health.Max = value },
		storeHistory: func(update *MonsterUpdateData, index int, value uint16) {
			update.HealthGraph103[index] = value
		},
		historyEndWord: func(update *MonsterUpdateData) uint16 {
			return uint16(uintptr(unsafe.Pointer(&update.HealthGraph103[0])) + unsafe.Sizeof(update.HealthGraph103))
		},
	}
}

// QuestHealthScale4E3DD0 scales only live, fully healed generators/monsters.
func (s *Server) QuestHealthScale4E3DD0(r QuestHealthScaleRuntime4E3DD0) int16 {
	return questHealthScale4E3DD0(s.questHealthScaleNativeHooks4E3DD0(r))
}
