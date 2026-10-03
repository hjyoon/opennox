package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageNPCHarpoonNativeCallback4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	var pin runtime.Pinner
	defer pin.Unpin()
	ud := &server.MonsterUpdateData{Field518: math.Float32bits(0.25), Field547: 99, Field546: 77}
	target := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(ud),
		HealthData: &server.HealthData{Cur: 60, Max: 60}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
	player := &server.Player{PlayerInd: 7}
	playerUD := &server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(0.125)}
	source := &server.Object{ObjClass: object.ClassPlayer | object.ClassComplex | object.ClassLight,
		UpdateData: unsafe.Pointer(playerUD), PrevPos: types.Ptf(44, 9)}
	// Stock NewPlayer/HarpoonBolt carry these non-layout class flags. The
	// bolt is MISSILE|WEAPON with ranged subclass 0x10, not a pure MISSILE.
	bolt := &server.Object{TypeInd: 66, ObjClass: object.ClassMissile | object.ClassWeapon | object.ClassComplex | object.ClassNotStackable,
		ObjSubClass: 0x10, ObjOwner: source, PrevPos: types.Ptf(20, 7), InitData: unsafe.Pointer(&server.ModifierInitData{})}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(bolt), target.UpdateData, source.UpdateData, bolt.InitData, unsafe.Pointer(player), unsafe.Pointer(target.HealthData)} {
		pin.Pin(pointer)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want above 4 GiB", pointer)
		}
	}
	beforeSource, beforeUD, beforePlayer, beforeBolt := *source, *playerUD, *player, *bolt
	for hit, want := range []struct {
		hp    uint16
		carry float32
	}{{58, 0.25}, {56, 0.5}, {53, -0.25}} {
		// Real C dispatcher and production PlayerDamage, DefaultDamage and
		// UnitSetHP services. No damage, HP, armor or collision substitution.
		// Stock WEAPON|MISSILE uses the owner's previous position for pain
		// direction, unlike the pure-MISSILE fixture's bolt position.
		if !objectDamageDispatchCallNative(target, source, bolt, 3, object.DamageImpact) || target.HealthData.Cur != want.hp ||
			math.Float32frombits(ud.Field1) != want.carry || ud.Field547 != 1 || ud.Field546 != uint32(bolt.TypeInd) ||
			target.Obj130 != bolt || target.Pos132 != source.PrevPos || target.Field131 != uint32(object.DamageImpact) || target.Frame134 != srv.Frame() ||
			*source != beforeSource || *playerUD != beforeUD || *player != beforePlayer || *bolt != beforeBolt {
			t.Fatalf("C entry harpoon=%d HP=%d carry=%g marker=%d/%d position=%v attacker=%p/%p type=%d frame=%d/%d player/UD/record/bolt preserved=%t/%t/%t/%t",
				hit, target.HealthData.Cur, math.Float32frombits(ud.Field1), ud.Field547, ud.Field546,
				target.Pos132, target.Obj130, bolt, target.Field131, target.Frame134, srv.Frame(),
				*source == beforeSource, *playerUD == beforeUD, *player == beforePlayer, *bolt == beforeBolt)
		}
	}
	t.Logf("C->NPC harpoon: target=%p player=%p bolt=%p HP=60->58->56->53 carry=%g", target, source, bolt, math.Float32frombits(ud.Field1))
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(bolt)
}
