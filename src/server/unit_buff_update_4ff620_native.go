package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

// UnitBuffUpdateRuntime4FF620 supplies the three effects whose implementation
// remains above package server or owns additional legacy restoration state.
// Every Object-bearing argument remains native-width.
type UnitBuffUpdateRuntime4FF620 struct {
	DamageClear        func(*Object, int32)
	IncrementElimDeath func(*Object)
	BuffOff            func(*Object, EnchantID)
}

type unitBuffUpdateNativeDeps4FF620 struct {
	loadUnitArg        func(*Object) *Object
	loadBuffs          func(*Object) uint32
	loadFPS            func() uint32
	loadDuration       func(*Object, int32) uint16
	storeDuration      func(*Object, int32, uint16)
	audio              func(int32, *Object, int32, int32)
	loadFlags          func(*Object) uint32
	storeFlags         func(*Object, uint32)
	storeObj130        func(*Object, *Object)
	storeDamageType    func(*Object, uint32)
	damageClear        func(*Object, int32)
	loadClassLow       func(*Object) uint8
	incrementElimDeath func(*Object)
	reportLesson       func(*Object)
	buffOff            func(*Object, int32) int32
	storePower         func(*Object, int32, uint8)
	testBuff           func(*Object, int32) int32
	loadSpeed          func(*Object) float32
	storeSpeed         func(*Object, float32)
}

func unitBuffUpdateNative4FF620(
	unit *Object,
	deps unitBuffUpdateNativeDeps4FF620,
) {
	UnitBuffUpdate4FF620(UnitBuffUpdateHooks4FF620[*Object]{
		LoadUnitArg: func() *Object {
			return deps.loadUnitArg(unit)
		},
		LoadBuffs:          deps.loadBuffs,
		LoadFPS:            deps.loadFPS,
		LoadDuration:       deps.loadDuration,
		StoreDuration:      deps.storeDuration,
		Audio:              deps.audio,
		LoadFlags:          deps.loadFlags,
		StoreFlags:         deps.storeFlags,
		StoreObj130:        deps.storeObj130,
		StoreDamageType:    deps.storeDamageType,
		DamageClear:        deps.damageClear,
		LoadClassLow:       deps.loadClassLow,
		IncrementElimDeath: deps.incrementElimDeath,
		ReportLesson:       deps.reportLesson,
		BuffOff:            deps.buffOff,
		StorePower:         deps.storePower,
		TestBuff:           deps.testBuff,
		LoadSpeed:          deps.loadSpeed,
		StoreSpeed:         deps.storeSpeed,
	})
}

func unitBuffUpdateServerDeps4FF620(
	s *Server,
	runtime UnitBuffUpdateRuntime4FF620,
) unitBuffUpdateNativeDeps4FF620 {
	return unitBuffUpdateNativeDeps4FF620{
		loadUnitArg: func(unit *Object) *Object {
			return unit
		},
		loadBuffs: func(unit *Object) uint32 {
			return unit.Buffs
		},
		loadFPS: s.TickRate,
		loadDuration: func(unit *Object, buff int32) uint16 {
			return unit.BuffsDur[int(buff)]
		},
		storeDuration: func(unit *Object, buff int32, value uint16) {
			unit.BuffsDur[int(buff)] = value
		},
		audio: func(id int32, unit *Object, kind, code int32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), uint32(code))
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		storeFlags: func(unit *Object, flags uint32) {
			unit.ObjFlags = object.Flags(flags)
		},
		storeObj130: func(unit, value *Object) {
			unit.Obj130 = value
		},
		storeDamageType: func(unit *Object, value uint32) {
			unit.Field131 = value
		},
		damageClear: runtime.DamageClear,
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		incrementElimDeath: runtime.IncrementElimDeath,
		reportLesson:       s.Nox_xxx_netReportLesson_4D8EF0,
		buffOff: func(unit *Object, buff int32) int32 {
			runtime.BuffOff(unit, EnchantID(buff))
			return 0
		},
		storePower: func(unit *Object, buff int32, value uint8) {
			unit.BuffsPower[int(buff)] = value
		},
		testBuff: func(unit *Object, buff int32) int32 {
			return unit.UnitBuffTest4FF350(buff)
		},
		loadSpeed: func(unit *Object) float32 {
			return unit.SpeedCur
		},
		storeSpeed: func(unit *Object, speed float32) {
			unit.SpeedCur = speed
		},
	}
}

// UnitBuffUpdate4FF620 binds GAME.EXE 004FF620 to native-width Object state.
// All fixed-width fields use their typed native layout; Obj130 and every
// callback object retain their complete pointer identity. A nil object faults
// at the original initial buff-dword load.
//
//go:noinline
func (s *Server) UnitBuffUpdate4FF620(
	unit *Object,
	runtime UnitBuffUpdateRuntime4FF620,
) {
	unitBuffUpdateNative4FF620(unit, unitBuffUpdateServerDeps4FF620(s, runtime))
}
