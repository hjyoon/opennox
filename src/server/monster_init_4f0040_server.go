package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterInitRuntime4F0040 supplies the one service still owned by the legacy
// runtime. All object, UpdateData, MonsterDef, HealthData, and AIStackItem
// pointers retain native width; SetHealth's amount remains the original low
// word produced through helper 00566DCC.
type MonsterInitRuntime4F0040 struct {
	SetHealth func(*Object, uint16)
}

func monsterInitNative4F0040(s *Server, unit *Object, runtime MonsterInitRuntime4F0040) {
	monsterInit4F0040(monsterInitHooks4F0040[
		*Object,
		*MonsterUpdateData,
		*HealthData,
		*MonsterDef,
		*AIStackItem,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(unit.UpdateData)
		},
		loadPlantTypeID: func() uint32 {
			return s.Types.monsterInitPlantID4F0040()
		},
		loadObjectFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadTypeIndex: func(unit *Object) uint16 {
			return unit.TypeInd
		},
		isRat: func(unit *Object) bool {
			ratTypeID := uint32(s.Types.RatID())
			return uint32(unit.TypeInd) == ratTypeID
		},
		isFish: func(unit *Object) bool {
			return s.Types.monsterInitIsFish4F0040(unit)
		},
		isFrog: func(unit *Object) bool {
			frogTypeID := uint32(s.Types.GreenFrogID())
			return uint32(unit.TypeInd) == frogTypeID
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		loadMonsterDef: func(update *MonsterUpdateData) *MonsterDef {
			return update.MonsterDef
		},
		loadMeleeRange: func(def *MonsterDef) float32 {
			return def.MeleeAttackRange112
		},
		loadCircleRadius: func(unit *Object) float32 {
			return unit.Shape.Circle.R
		},
		loadAIAction: func(update *MonsterUpdateData) uint32 {
			return update.AIAction340
		},
		storeAIAction: func(update *MonsterUpdateData, action uint32) {
			update.AIAction340 = action
		},
		storeSightRange: func(update *MonsterUpdateData, value float32) {
			update.SightRange = value
		},
		storeAggression: func(update *MonsterUpdateData, value float32) {
			update.Aggression = value
		},
		loadStatus: func(update *MonsterUpdateData) uint32 {
			return uint32(update.StatusFlags)
		},
		storeStatus: func(update *MonsterUpdateData, status uint32) {
			update.StatusFlags = objectMonsterStatus4F0040(status)
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeActionArg: func(item *AIStackItem, index int, value uint32) {
			item.Args[index] = uintptr(value)
		},
		storeActionArgLow: func(item *AIStackItem, index int, value uint8) {
			item.Args[index] = item.Args[index]&^uintptr(0xff) | uintptr(value)
		},
		canAttackAtWill: func(unit *Object) bool {
			return (*MonsterUpdateData)(unit.UpdateData).Aggression > 0.66000003
		},
		loadPositionXBits: func(unit *Object) uint32 {
			return math.Float32bits(unit.PosVec.X)
		},
		loadPositionYBits: func(unit *Object) uint32 {
			return math.Float32bits(unit.PosVec.Y)
		},
		loadFrame: func() uint32 {
			return s.Frame()
		},
		loadDirection: func(unit *Object) int16 {
			return int16(unit.Direction1)
		},
		storeDirection: func(update *MonsterUpdateData, value uint32) {
			update.Direction94 = value
		},
		storePositionX: func(update *MonsterUpdateData, value uint32) {
			update.Pos95.X = math.Float32frombits(value)
		},
		storePositionY: func(update *MonsterUpdateData, value uint32) {
			update.Pos95.Y = math.Float32frombits(value)
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadHealthMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		loadHealthCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadHealthScale: func(update *MonsterUpdateData) float32 {
			return update.Field338
		},
		setHealth: runtime.SetHealth,
		storeHealthPrev: func(health *HealthData, value uint16) {
			health.Field2 = value
		},
		storeHealthGraph: func(update *MonsterUpdateData, index int, value uint16) {
			update.HealthGraph103[index] = value
		},
		loadSubclassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjSubClass)
		},
		loadSpeedField332: func(update *MonsterUpdateData) float32 {
			return update.Field332
		},
		loadSpeedField333: func(update *MonsterUpdateData) uint32 {
			return update.Field333
		},
		randomFloat: func(minimum, maximum float32, _ string, _ int) float64 {
			return logicRandomFloat416030(s.Rand.Logic, minimum, maximum)
		},
		loadSpeedBase: func(unit *Object) float32 {
			return unit.SpeedBase
		},
		storeSpeedBase: func(unit *Object, value float32) {
			unit.SpeedBase = value
		},
		canCast: func(unit *Object) bool {
			return uint32((*MonsterUpdateData)(unit.UpdateData).StatusFlags)&0x20 != 0
		},
		storeFleeRange: func(update *MonsterUpdateData, value float32) {
			update.FleeRange = value
		},
	})
}

// MonsterInit4F0040 binds GAME.EXE 004F0040 to native-width server storage
// without changing the original fixed-width fields or observable order.
func (s *Server) MonsterInit4F0040(unit *Object, runtime MonsterInitRuntime4F0040) {
	monsterInitNative4F0040(s, unit, runtime)
}

// Keep the conversion isolated so all status reads and writes remain exact
// fixed-width dwords without importing object into the semantic core.
func objectMonsterStatus4F0040(value uint32) object.MonsterStatus {
	return object.MonsterStatus(value)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.Direction94)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.AIAction340)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.StatusFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(AIStackItem{}.Action)]
	_ = [1]struct{}{}[2-unsafe.Sizeof(HealthData{}.Cur)]
)
