package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// These are direct registered PlayerDamage calls, including raw-1 ABI
// compatibility. The public DoDamage(DamageTrue) bypass is a separate,
// unchanged caller and is not exercised by this callback matrix.
func TestSourceLessPlayerAllTypesNativeCallback4E17B0(t *testing.T) {
	_ = npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold source-less hit requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	var damageTypes []object.DamageType
	for typ := object.DamageTrue; typ <= object.DamageAirborneElectric; typ++ {
		damageTypes = append(damageTypes, typ)
	}
	damageTypes = append(damageTypes, object.DamageType(math.MinInt32), -2, 18, 19, object.DamageType(math.MaxInt32))
	for _, typ := range damageTypes {
		for _, god := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/god-%t", int32(typ), god), func(t *testing.T) {
				if god {
					noxflags.SetEngine(noxflags.EngineGodMode)
				} else {
					noxflags.UnsetEngine(noxflags.EngineGodMode)
				}
				var pin runtime.Pinner
				defer pin.Unpin()
				npc, target := npcReflectLegacyObjects4E17B0(t, &pin)
				target.Damage = playerDamageMeleeCallbackNative4E17B0()
				target.Field38, target.Obj130, target.Frame134 = 0, npc, 77
				target.Pos132 = types.Ptf(29, 31)
				ud := target.UpdateDataPlayer()
				ud.Field57, ud.Field21 = math.Float32bits(.5), math.Float32bits(.25)
				ud.Field76, ud.Field75 = 99, 77
				ud.Field40_0, ud.Field40_1 = 0x1122, 0xabcd
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
					t.Fatal("source-less player must retain a native high address")
				}
				wantHP, wantCarry := uint16(191), float32(.25)
				switch typ {
				case object.DamageBlade, object.DamageImpale, object.DamageExplosion, object.DamageBite, object.DamageClaw, object.DamageImpact:
					wantHP, wantCarry = 196, .75
				case object.DamageCrush:
					wantHP, wantCarry = 193, 0
				}
				if god {
					wantHP = 200
				}
				wantMarker, wantKind := uint32(2), uint32(typ)
				if typ < object.DamageBlade || typ > object.DamageAirborneElectric || typ == object.DamageDeathMagic {
					wantMarker, wantKind = 0, 77
				}
				if !objectDamageDispatchCallNative(target, nil, nil, 9, typ) || target.HealthData.Cur != wantHP || ud.Field21 != math.Float32bits(wantCarry) || ud.Field76 != wantMarker || ud.Field75 != wantKind {
					t.Fatalf("C->PlayerDamage->DefaultDamage->UnitSetHP type=%d god=%t HP=%d/%d carry=%x marker=%d/%d", int32(typ), god, target.HealthData.Cur, wantHP, ud.Field21, ud.Field76, ud.Field75)
				}
				if !god && (target.Field38 != math.MaxUint32 || target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 1400 || target.Field131 != uint32(typ)) {
					t.Fatal("source-less HP attribution/sync changed")
				}
				if god && (target.Field38 != 0 || target.Obj130 != npc || target.Pos132 != types.Ptf(29, 31) || target.Frame134 != 77) {
					t.Fatal("GodMode crossed the HP boundary")
				}
				wantElectric := uint16(0x1122)
				if !god && (typ == object.DamageElectric || typ == object.DamageAirborneElectric) {
					wantElectric = 2
				}
				if ud.Field40_0 != wantElectric || ud.Field40_1 != 0xabcd {
					t.Fatal("electric WORD state or its adjacent WORD changed")
				}
				runtime.KeepAlive(npc)
				runtime.KeepAlive(target)
			})
		}
	}
}
