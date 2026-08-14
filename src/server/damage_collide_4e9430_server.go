package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// DamageCollideData is the pointer-independent eight-byte collide record
// registered for DamageCollide. GAME.EXE 004E9430 reads Damage and DamageType;
// its parser deliberately leaves Reserved untouched.
type DamageCollideData struct {
	Damage     uint8
	Reserved   [3]uint8
	DamageType int32
}

type damageCollideNativeDeps4E9430 struct {
	loadFrameLow func() uint8
	findParent   func(*Object) *Object
	damage       func(*Object, *Object, *Object, int32, int32) int32
}

func damageCollideNative4E9430(
	source, target *Object,
	collision unsafe.Pointer,
	deps damageCollideNativeDeps4E9430,
) {
	damageCollide4E9430(source, target, collision, damageCollideHooks4E9430[
		*Object,
		*DamageCollideData,
		*HealthData,
	]{
		loadCollideData: func(obj *Object) *DamageCollideData {
			return (*DamageCollideData)(obj.CollideData)
		},
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		loadDamage: func(data *DamageCollideData) uint8 {
			return data.Damage
		},
		loadFrameLow: deps.loadFrameLow,
		loadDamageType: func(data *DamageCollideData) int32 {
			return data.DamageType
		},
		findParent: deps.findParent,
		damage:     deps.damage,
	})
}

// DamageCollide4E9430 binds the original DamageCollide callback to native
// Object, HealthData, DamageCollideData and damage-function pointer widths.
func (s *Server) DamageCollide4E9430(source, target *Object, collision unsafe.Pointer) {
	damageCollideNative4E9430(source, target, collision, damageCollideNativeDeps4E9430{
		loadFrameLow: func() uint8 {
			return uint8(s.Frame())
		},
		findParent: (*Object).FindOwnerChainPlayer,
		damage: func(target, source, attacker *Object, damage, damageType int32) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(source)),
				uintptr(toObjectC(attacker)),
				uintptr(uint32(damage)),
				uintptr(uint32(damageType)),
			))
		},
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(DamageCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(DamageCollideData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(DamageCollideData{}.DamageType)]
)
