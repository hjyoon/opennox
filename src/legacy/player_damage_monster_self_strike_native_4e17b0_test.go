package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// These pinned native-width records traverse the real C damage dispatcher,
// registered PlayerDamage, DefaultDamage and UnitSetHP, without replacing HP.
func TestPlayerDamageMonsterSelfStrikeNative4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold strike requested hurt-state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageDrain, object.DamageClaw} {
		for _, absorption := range []float32{0, 0.25, 1} {
			t.Run(fmt.Sprintf("type-%d/armor-%g", typ, absorption), func(t *testing.T) {
				player := &server.Player{}
				ud := &server.PlayerUpdateData{Player: player, State: server.PlayerState13, Field57: math.Float32bits(absorption), Field21: math.Float32bits(0.125), Field76: 99, Field75: 77}
				attackerUD := &server.MonsterUpdateData{}
				v := &server.Object{TypeInd: 71, ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud), HealthData: &server.HealthData{Cur: 200, Field2: 200, Max: 200}, Material: 0x4000, Damage: playerDamageMeleeCallbackNative4E17B0()}
				a := &server.Object{TypeInd: 713, ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(attackerUD), HealthData: &server.HealthData{Cur: 200, Max: 200}, Material: 0x4000, PrevPos: types.Ptf(-20, 7)}
				var pin runtime.Pinner
				defer pin.Unpin()
				for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(player), unsafe.Pointer(ud), unsafe.Pointer(attackerUD), unsafe.Pointer(v.HealthData), unsafe.Pointer(a.HealthData)} {
					pin.Pin(p)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
						t.Fatalf("pointer=%p, want >4 GiB", p)
					}
				}
				x := float64(absorption)
				if typ == object.DamageCrush {
					x *= 0.5
				}
				accumulated := float32((1-x)*10) + 0.125
				rounded := int32(math.RoundToEven(float64(accumulated)))
				effective := max(rounded, 1)
				carry := accumulated - float32(rounded)
				if typ == object.DamageDrain {
					effective, carry = 10, 0.125
				}
				if !srv.IsEnemyTo(v, a) {
					t.Fatal("native self-strike fixture is not hostile")
				}
				if !objectDamageDispatchCallNative(v, a, a, 10, typ) || v.HealthData.Cur != uint16(200-effective) || v.Field38 != math.MaxUint32 || ud.Field21 != math.Float32bits(carry) || ud.Field76 != 2 || ud.Field75 != uint32(typ) || v.Obj130 != a || v.Field131 != uint32(typ) || v.Frame134 != 1400 || attackerUD.Field130 != 1400 {
					t.Fatalf("native self-strike: HP=%d want=%d marker=%d/%d", v.HealthData.Cur, 200-effective, ud.Field76, ud.Field75)
				}
				if a.HealthData.Cur != 200 {
					t.Fatal("strike damaged the attacker")
				}
				t.Logf("C->PlayerDamage->DefaultDamage->UnitSetHP: target=%p source=weapon=%p type=%d HP=200->%d", v, a, typ, v.HealthData.Cur)
			})
		}
	}
}
