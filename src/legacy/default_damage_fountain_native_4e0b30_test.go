package legacy

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// These synthetic records retain the exact reported demon and fountain
// type/class/subclass/flags. All C-dispatch records are C-owned above 4 GiB.
// Registered production damage callbacks and UnitSetHP are not substituted.
func TestDefaultDamageFountainNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeFountainDefault"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(64) {
		t.Fatal("native fountain object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, target := range []struct {
		name       string
		typ        uint16
		class      object.Class
		subclass   object.SubClass
		flags      object.Flags
		playerPath bool
	}{
		{"MeleeDemon", 1338, 2621442, 332802, 554172932, false},
		{"EmberDemon", 1337, 2621442, 332802, 554172932, false},
		{"ordinary-monster", 2273, 2097154, 66050, 16777732, false},
		{"NPC-DefaultDamage", 1343, 2097154, 0x11012, 16777732, false},
		{"NPC-PlayerDamage", 1343, 2097154, 0x11012, 16777732, true},
		{"player", 713, 2621444, 0, 16777732, true},
	} {
		for _, immune := range []bool{false, true} {
			if target.name == "player" && immune {
				continue // The 0x400 immunity is a monster subclass bit.
			}
			t.Run(fmt.Sprintf("%s/fire-immune-%t", target.name, immune), func(t *testing.T) {
				var diagnostics bytes.Buffer
				srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
				v, a, w := spellMissileImpactCRecords4E17B0(t, srv, "self")
				if a != w {
					t.Fatal("native fountain source/weapon identity")
				}
				v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = target.typ, target.class, target.subclass, target.flags
				v.Frame134, v.Pos132, v.Buffs = 77, types.Ptf(31, 47), 0
				v.Damage = srv.Types.ByID(id).Damage
				if target.playerPath {
					v.Damage = playerDamageMeleeCallbackNative4E17B0()
				}
				w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 1202, 4259840, 0, 16777748
				w.PrevPos, w.ObjOwner = types.Ptf(17, 19), nil
				var carry, marker, markerType *uint32
				var ud *server.MonsterUpdateData
				if v.Class().Has(object.ClassPlayer) {
					u := v.UpdateDataPlayer()
					u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 91
					carry, marker, markerType = &u.Field21, &u.Field76, &u.Field75
				} else {
					u, freeU := alloc.New(server.MonsterUpdateData{})
					t.Cleanup(freeU)
					*u = server.MonsterUpdateData{Field547: 99, Field546: 91, Field523_2: 88, Field518: math.Float32bits(.5), Field1: math.Float32bits(.25)}
					v.UpdateData, ud = unsafe.Pointer(u), u
					carry, marker, markerType = &u.Field1, &u.Field547, &u.Field546
					v.ObjSubClass &^= 0x400
					if immune {
						v.ObjSubClass |= 0x400
					}
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(v.UpdateData) <= math.MaxUint32 {
						t.Fatal("native monster update is not C-owned above 4 GiB")
					}
				}
				beforeW, beforeSync := *w, v.Field38
				hp := []uint16{v.HealthData.Cur}
				for hit := 1; hit <= 3; hit++ {
					wantHP, wantMarker, wantType := uint16(60-hit*3), uint32(2), uint32(1)
					if immune {
						wantHP, wantMarker = 60, 0
						if !target.playerPath {
							wantType = 91
						}
					}
					result := objectDamageDispatchCallNative(v, w, w, 3, object.DamageFlame)
					hp = append(hp, v.HealthData.Cur)
					if strings.Contains(diagnostics.String(), "native branch is not ported") {
						t.Fatalf("reported shape remains unsupported: %s", diagnostics.String())
					}
					if !result || v.HealthData.Cur != wantHP || v.HealthData.Max != 60 || v.HealthData.Field2 != 60 ||
						*carry != math.Float32bits(.25) || *marker != wantMarker || *markerType != wantType || *w != beforeW {
						t.Fatalf("native fountain hit=%d HP=%d/%d marker=%d/%d type=%d/%d result=%t", hit, v.HealthData.Cur, wantHP, *marker, wantMarker, *markerType, wantType, result)
					}
					if immune {
						if v.Obj130 != nil || v.Frame134 != 77 || v.Field131 != 0 || v.Pos132 != types.Ptf(31, 47) || v.Field38 != beforeSync {
							t.Fatal("immune fountain hit entered HP attribution/synchronization")
						}
					} else if v.Obj130 != w || v.Frame134 != 1400 || v.Field131 != 1 || v.Pos132 != w.PrevPos || v.Field38 != math.MaxUint32 {
						t.Fatal("native fountain lost HP attribution/synchronization")
					}
					if ud != nil && (!ud.StatusFlags.Has(object.MonStatusOnFire) || ud.StatusFlags.Has(object.MonStatusInjured) == immune ||
						ud.Field1 != math.Float32bits(.25) || ud.Field518 != math.Float32bits(.5) || ud.Field523_2 != 88) {
						t.Fatal("native fountain monster status/armor/carry changed")
					}
					if ud == nil && (v.UpdateDataPlayer().State != server.PlayerState13 || v.UpdateDataPlayer().Field57 != math.Float32bits(.5)) {
						t.Fatal("small world FLAME changed player hurt state/absorption")
					}
				}
				t.Logf("C-owned >4 GiB C->registered damage->DefaultDamage->UnitSetHP: target=%s type=%d class=%d subclass=%d flags=%d source=weapon=%p source_type=1202 source_class=4259840 source_flags=16777748 damage=3 damage_type=1 immune=%t HP=%v", target.name, v.TypeInd, uint32(v.ObjClass), uint32(v.ObjSubClass), uint32(v.ObjFlags), w, immune, hp)
			})
		}
	}
}
