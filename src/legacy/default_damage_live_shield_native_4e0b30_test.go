package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Direct Go damage body and production runtime, not an unmodified C damage
// dispatcher or a naturally changing stock-map Shield. Only the real BuffOff
// service is wrapped to add/remove Shield through the production buff APIs.
// The duration head is an injected C-owned fixture; reduction, duration walk,
// network FX and UnitSetHP use their unchanged production implementations.
func TestDefaultDamageLiveShieldNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(9) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, initial := range []bool{false, true} {
		t.Run(fmt.Sprintf("initial-%t", initial), func(t *testing.T) {
			target, source, weapon := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
			update, freeUpdate := alloc.New(server.MonsterUpdateData{})
			playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			health, freeHealth := alloc.New(server.HealthData{})
			initData, freeInit := alloc.New(server.ModifierInitData{})
			duration, freeDuration := alloc.New(server.DurSpell{})
			for _, free := range []func(){freeUpdate, freePlayerUpdate, freePlayer, freeHealth, freeInit, freeDuration} {
				t.Cleanup(free)
			}
			*player = server.Player{PlayerInd: 7}
			*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			target.TypeInd, target.ObjClass, target.ObjFlags, target.ObjSubClass = 71, object.ClassMonster, 0, 0x10
			target.UpdateData, target.HealthData = unsafe.Pointer(update), health
			source.TypeInd, source.ObjClass, source.ObjFlags = 72, object.ClassPlayer, 0
			source.UpdateData, source.PrevPos = unsafe.Pointer(playerUpdate), types.Ptf(-3, 7)
			weapon.TypeInd, weapon.ObjClass, weapon.ObjFlags, weapon.ObjSubClass = 777, object.ClassWand, 0, 0
			weapon.InitData = unsafe.Pointer(initData)
			*duration = server.DurSpell{Spell: uint32(spell.SPELL_SHIELD), Target48: target, Level: 1, Field72: 42}
			oldHead := srv.Spells.Dur.List
			srv.Spells.Dur.List = duration
			t.Cleanup(func() { srv.Spells.Dur.List = oldHead }) // Before record free/Server.Close.
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(update), unsafe.Pointer(playerUpdate), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(initData), unsafe.Pointer(duration)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want >4 GiB", ptr)
				}
			}
			if initial {
				buffApplyExportCall4FF380(target, int32(server.ENCHANT_SHIELD), 600, 1)
			}
			r := defaultDamageWorldRuntime4E0B30(srv)
			buffOff, calls := r.BuffOff, 0
			r.BuffOff = func(owner *server.Object, id server.EnchantID) {
				if owner != target || id != server.ENCHANT_INVISIBLE || update.Field547 != 1 || update.Field546 != 777 ||
					target.Pos132 != source.PrevPos || health.Cur != 20 || duration.Field72 != 42 {
					t.Fatal("Shield mutation ran outside the original pre-attribution hit prefix")
				}
				buffOff(owner, id)
				calls++
				if initial {
					Nox_xxx_spellBuffOff_4FF5B0(owner, server.ENCHANT_SHIELD)
				} else {
					buffApplyExportCall4FF380(owner, int32(server.ENCHANT_SHIELD), 600, 1)
				}
			}
			wantHP, wantShieldHP := uint16(16), int32(38)
			if initial {
				wantHP, wantShieldHP = 12, 42
			}
			if !server.DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageBlade, r) || calls != 1 || health.Cur != wantHP || duration.Field72 != wantShieldHP ||
				target.HasEnchant(server.ENCHANT_SHIELD) == initial || target.Obj130 != weapon || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageBlade) ||
				target.Field38 != math.MaxUint32 || update.Field547 != 1 || update.Field546 != 777 || !update.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatalf("native live Shield: initial=%t live=%t HP=%d want=%d duration=%d want=%d mutations=%d", initial, target.HasEnchant(server.ENCHANT_SHIELD), health.Cur, wantHP, duration.Field72, wantShieldHP, calls)
			}
			t.Logf("C-owned >4GiB production Shield: target=%p duration=%p initial=%t live=%t HP=20->%d Shield=42->%d", target, duration, initial, !initial, health.Cur, duration.Field72)
		})
	}
}
